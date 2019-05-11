package runner

// verify should only return an error if there's an error that should stop execution
// for transient errors, return false, nil and expect a retry/backoff interval
func verify(cluster *Cluster, verification *TestVerification) (bool, []byte, []byte, error) {
	exitCode, stdout, stderr, err := cluster.exec(verification.Exec.Pod, verification.Exec.Command, verification.Exec.Args)
	if err != nil {
		return false, nil, nil, err
	}

	return exitCode == 0, stdout, stderr, nil
}
