package main

import (
	"github.com/schemahero/schemahero/pkg/cli/schemaherokubectlcli"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	schemaherokubectlcli.InitAndExecute()
}
