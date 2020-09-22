package config

import (
	flag "github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

var (
	kubernetesConfigFlags *genericclioptions.ConfigFlags
)

func init() {
	kubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
}

func AddFlags(flags *flag.FlagSet) {
	kubernetesConfigFlags.AddFlags(flags)
}

func GetRESTConfig() (*rest.Config, error) {
	return kubernetesConfigFlags.ToRESTConfig()
}
