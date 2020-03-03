package schemaherokubectlcli

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/schemahero/schemahero/pkg/installer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func InstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "install",
		Short:         "install the schemahero operator to the cluster",
		Long:          `...`,
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()

			if v.GetString("extensions-api") != "" {
				if v.GetString("extensions-api") != "v1" && v.GetString("extensions-api") != "v1beta1" {
					fmt.Printf("Unsupported value in extensions-api %q, only v1 and v1beta1 are supported\n", v.GetString("extensions-api"))
					os.Exit(1)
				}
			}

			if v.GetBool("yaml") {
				manifests, err := installer.GenerateOperatorYAML(v.GetString("extensions-api"), v.GetBool("enterprise"), v.GetString("enterprise-tag"), v.GetString("namespace"))
				if err != nil {
					fmt.Printf("Error: %s\n", err.Error())
					return err
				}

				if v.GetString("out-dir") != "" {
					if err := os.MkdirAll(v.GetString("out-dir"), 0755); err != nil {
						fmt.Printf("Error: %s\n", err.Error())
						return err
					}

					for filename, manifest := range manifests {
						if err := ioutil.WriteFile(filepath.Join(v.GetString("out-dir"), filename), manifest, 0644); err != nil {
							fmt.Printf("Error: %s\n", err.Error())
							return err
						}
					}
				} else {
					// Write to stdout
					multiDocResult := [][]byte{}
					for _, manifest := range manifests {
						multiDocResult = append(multiDocResult, manifest)
					}

					fmt.Println(string(bytes.Join(multiDocResult, []byte("\n---\n"))))
				}
				return nil
			}
			if err := installer.InstallOperator(v.GetBool("enterprise"), v.GetString("namespace")); err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				return err
			}

			fmt.Println("The SchemaHero operator has been installed to the cluster")
			return nil
		},
	}

	cmd.Flags().Bool("yaml", false, "If present, don't install the operator, just generate the yaml")
	cmd.Flags().String("out-dir", "", "If present and --yaml also specified, write all of the manifests to this directory")
	cmd.Flags().String("extensions-api", "", "version of apiextensions.k8s.io to generate. if unset, will detect best version from kubernetes version")
	cmd.Flags().Bool("enterprise", false, "If preset, generate enterprise YAML with KOTS template functions. This probably isn't what you want")
	cmd.Flags().String("enterprise-tag", "latest", "the tag of the enterprise images to include")
	cmd.Flags().StringP("namespace", "n", "schemahero-system", "The namespace to install SchemaHero Operator into")

	return cmd
}
