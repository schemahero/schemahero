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
		Use:   "install",
		Short: "install or upgrade the schemahero operator in a cluster (or generate the yaml)",
		Long: `The install command will install SchemaHero into a cluster, or upgrade an existing installation if one is found.

When upgrading, the command will replace the SchemaHero operator with the latest and will likely overwrite any manual changes to the operator manifest.
After upgrading, the operator will roll out new database managers, restarting each with the newest version.

For more control, use the --yaml flag to avoid making any changes to the cluster, and only print the manifests to the terminal that can be deployed using other tooling.`,
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()

			if v.GetBool("yaml") {
				manifests, err := installer.GenerateOperatorYAML(v.GetString("namespace"))
				if err != nil {
					fmt.Printf("Error: %s\n", err.Error())
					return err
				}

				if v.GetString("out-dir") != "" {
					if err := os.MkdirAll(v.GetString("out-dir"), 0750); err != nil {
						fmt.Printf("Error: %s\n", err.Error())
						return err
					}

					for filename, manifest := range manifests {
						if err := ioutil.WriteFile(filepath.Join(v.GetString("out-dir"), filename), manifest, 0600); err != nil {
							fmt.Printf("Error: %s\n", err.Error())
							return err
						}
					}
				} else {
					// Write to stdout

					// write the namespace first
					if namespace, ok := manifests["namespace.yaml"]; ok {
						fmt.Printf("%s---\n", namespace)
					}

					delete(manifests, "namespace.yaml")

					multiDocResult := [][]byte{}
					for _, manifest := range manifests {
						multiDocResult = append(multiDocResult, manifest)
					}

					fmt.Println(string(bytes.Join(multiDocResult, []byte("\n---\n"))))
				}
				return nil
			}

			wasUpgraded, err := installer.InstallOperator(v.GetString("namespace"))
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				return err
			}

			if !wasUpgraded {
				fmt.Println("The SchemaHero operator has been installed to the cluster")
			} else {
				fmt.Println("The SchemaHero operator has been upgraded in the cluster")
			}

			return nil
		},
	}

	cmd.Flags().Bool("yaml", false, "If present, don't install the operator, just generate the yaml")
	cmd.Flags().String("out-dir", "", "If present and --yaml also specified, write all of the manifests to this directory")

	cmd.Flags().StringP("namespace", "n", "schemahero-system", "The namespace to install SchemaHero Operator into")

	return cmd
}
