#!/usr/bin/env python3
"""Summarize results from local validation checks that wrote logs to a temp dir.

Usage:
    python3 scripts/summarize_local_checks.py <tmpdir> <name> <exit_code> ...

For each name, reads <tmpdir>/<name>.log and emits a JSON summary with the
overall status, per-check status, and a tail of the log file.
"""
import json
import pathlib
import sys


def main():
    if len(sys.argv) < 4 or (len(sys.argv) - 2) % 2 != 0:
        print(
            "Usage: summarize_local_checks.py <tmpdir> <name> <exit_code> ...",
            file=sys.stderr,
        )
        sys.exit(1)

    tmpdir = pathlib.Path(sys.argv[1])
    pairs = sys.argv[2:]
    statuses = {}
    for i in range(0, len(pairs), 2):
        name = pairs[i]
        try:
            code = int(pairs[i + 1])
        except ValueError as e:
            print(f"Invalid exit code for {name}: {pairs[i + 1]}", file=sys.stderr)
            sys.exit(1)
        statuses[name] = code

    checks = []
    for name, code in statuses.items():
        log_path = tmpdir / f"{name}.log"
        log_text = log_path.read_text(errors="replace") if log_path.exists() else ""
        checks.append({
            "name": name,
            "status": "passed" if code == 0 else "failed",
            "exit_code": code,
            "log_tail": log_text[-12000:],
        })

    failed = [check for check in checks if check["exit_code"] != 0]
    print(json.dumps({
        "status": "passed" if not failed else "failed",
        "reason": "all validation checks passed" if not failed else "one or more validation checks failed",
        "checks": checks,
    }))


if __name__ == "__main__":
    main()
