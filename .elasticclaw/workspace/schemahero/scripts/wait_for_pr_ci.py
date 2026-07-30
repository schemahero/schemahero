#!/usr/bin/env python3
"""Poll a PR's GitHub checks, retry failed GitHub Actions jobs once, and surface logs.

Usage:
    python3 scripts/wait_for_pr_ci.py \
        --repo-dir <repo-dir> \
        --branch <branch> \
        [--max-wait <seconds>] \
        [--poll-interval <seconds>] \
        [--progress-interval <seconds>]

Outputs JSON with status, pr_url, checks, retried runs, and failed log tails.
"""
import argparse
import json
import os
import subprocess
import sys
import time

OK_STATES = {"success", "skipped", "neutral"}
PENDING_STATES = {
    "pending", "expected", "queued", "in_progress", "waiting", "action_required", "stale"
}
FAILED_STATES = {"failure", "cancelled", "timed_out"}


def normalize_state(state):
    return state.lower() if state else state


def run_gh(args, check=True, cwd=None):
    env = os.environ.copy()
    env["GH_PAGER"] = "cat"
    env["GH_PROMPT_DISABLED"] = "1"
    env["NO_COLOR"] = "1"
    return subprocess.run(
        ["gh"] + args,
        capture_output=True,
        text=True,
        check=check,
        cwd=cwd,
        env=env,
    )


def get_repo_name(repo_dir):
    result = run_gh(
        ["repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner"],
        cwd=repo_dir,
    )
    return result.stdout.strip()


def get_pr_info(repo_name, branch):
    result = run_gh([
        "pr", "list",
        f"--repo={repo_name}",
        f"--head={branch}",
        "--json", "number,url",
        "--jq", ".[0]",
    ])
    data = json.loads(result.stdout)
    if not data:
        raise RuntimeError(f"no open PR found for branch {branch}")
    return data


def get_pr_checks(repo_name, pr_number):
    # gh pr checks exits non-zero when some checks failed, so do not check return code.
    result = run_gh([
        "pr", "checks", str(pr_number),
        f"--repo={repo_name}",
        "--json", "name,state,link",
    ], check=False)
    return json.loads(result.stdout) if result.stdout else []


def wait_for_checks(repo_name, pr_number, max_wait_seconds, poll_interval, progress_interval):
    start = time.time()
    last_progress = start
    while True:
        checks = get_pr_checks(repo_name, pr_number)
        pending = [c for c in checks if normalize_state(c.get("state")) in PENDING_STATES]
        if not pending:
            return checks
        now = time.time()
        if now - last_progress >= progress_interval:
            elapsed = now - start
            pending_names = [c.get("name") for c in pending]
            print(
                f"Waiting for PR checks... elapsed: {elapsed:.0f}s, pending: {len(pending)} ({', '.join(pending_names)})",
                file=sys.stderr,
            )
            last_progress = now
        if now - start > max_wait_seconds:
            raise TimeoutError("checks did not complete within the time limit")
        time.sleep(poll_interval)


def rerun_failed_runs(repo_name, branch):
    result = run_gh([
        "run", "list",
        f"--repo={repo_name}",
        f"--branch={branch}",
        "--status", "failure",
        "--json", "databaseId,displayTitle",
    ])
    runs = json.loads(result.stdout)
    retried = []
    for run in runs:
        run_id = run["databaseId"]
        try:
            run_gh(["run", "rerun", f"--repo={repo_name}", "--failed", str(run_id)])
            retried.append({
                "run_id": run_id,
                "display_title": run.get("displayTitle"),
                "rerun_triggered": True,
            })
        except subprocess.CalledProcessError as e:
            retried.append({
                "run_id": run_id,
                "display_title": run.get("displayTitle"),
                "rerun_triggered": False,
                "error": (e.stderr or e.stdout),
            })
    return retried


def get_failed_logs(repo_name, retried):
    logs = []
    for item in retried:
        if not item.get("rerun_triggered"):
            continue
        run_id = item["run_id"]
        try:
            result = run_gh(["run", "view", f"--repo={repo_name}", "--log-failed", str(run_id)])
            log_text = result.stdout
            logs.append({"run_id": run_id, "log_tail": log_text[-12000:]})
        except subprocess.CalledProcessError as e:
            logs.append({"run_id": run_id, "error": (e.stderr or e.stdout)})
    return logs


def main():
    parser = argparse.ArgumentParser(
        description="Wait for PR checks and retry failed GitHub Actions jobs once.",
    )
    parser.add_argument("--repo-dir", default=".", help="Path to the repository checkout")
    parser.add_argument("--branch", required=True, help="Branch name for the PR")
    parser.add_argument("--max-wait", type=int, default=3600, help="Maximum seconds to wait for checks")
    parser.add_argument("--poll-interval", type=int, default=30, help="Seconds between polls")
    parser.add_argument("--progress-interval", type=int, default=300, help="Seconds between progress updates")
    args = parser.parse_args()

    if not os.path.isdir(args.repo_dir):
        raise RuntimeError(f"repo directory does not exist: {args.repo_dir}")

    repo_name = get_repo_name(args.repo_dir)
    pr_info = get_pr_info(repo_name, args.branch)
    pr_number = pr_info.get("number")
    pr_url = pr_info.get("url")

    checks = wait_for_checks(repo_name, pr_number, args.max_wait, args.poll_interval, args.progress_interval)
    failed = [c for c in checks if normalize_state(c.get("state")) in FAILED_STATES]

    if not failed:
        print(json.dumps({"status": "passed", "pr_url": pr_url, "checks": checks}))
        return

    retried = rerun_failed_runs(repo_name, args.branch)
    checks = wait_for_checks(repo_name, pr_number, args.max_wait, args.poll_interval, args.progress_interval)
    failed = [c for c in checks if normalize_state(c.get("state")) in FAILED_STATES]

    if not failed:
        print(json.dumps({
            "status": "passed",
            "pr_url": pr_url,
            "checks": checks,
            "retried": retried,
        }))
        return

    logs = get_failed_logs(repo_name, retried)
    print(json.dumps({
        "status": "failed",
        "pr_url": pr_url,
        "checks": checks,
        "retried": retried,
        "logs": logs,
    }))


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(json.dumps({"status": "error", "reason": str(e)}))
        sys.exit(0)
