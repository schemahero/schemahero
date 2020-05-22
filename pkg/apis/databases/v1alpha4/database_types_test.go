package v1alpha3

import (
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestVaultConnectionURI(t *testing.T) {
	tests := []struct {
		name string
		db   *Database
		want string
	}{
		{
			name: "Postgres",
			db: &Database{
				ObjectMeta: v1.ObjectMeta{
					Name: "testdb",
				},
				Spec: DatabaseSpec{
					Connection: DatabaseConnection{
						Postgres: &PostgresConnection{
							URI: ValueOrValueFrom{
								ValueFrom: &ValueFrom{
									Vault: &Vault{
										Secret: "database/creds/test",
									},
								},
							},
						},
					},
				},
			},
			want: `
{{- with secret "database/creds/test" -}}
postgres://{{ .Data.username }}:{{ .Data.password }}@postgres:5432/testdb{{- end }}`,
		},
		{
			name: "MySQL",
			db: &Database{
				ObjectMeta: v1.ObjectMeta{
					Name: "testdb",
				},
				Spec: DatabaseSpec{
					Connection: DatabaseConnection{
						Mysql: &MysqlConnection{
							URI: ValueOrValueFrom{
								ValueFrom: &ValueFrom{
									Vault: &Vault{
										Secret: "database/creds/test",
									},
								},
							},
						},
					},
				},
			},
			want: `
{{- with secret "database/creds/test" -}}
mysql://{{ .Data.username }}:{{ .Data.password }}@mysql:3306/testdb{{- end }}`,
		},
		{
			name: "CockroachDB",
			db: &Database{
				ObjectMeta: v1.ObjectMeta{
					Name: "testdb",
				},
				Spec: DatabaseSpec{
					Connection: DatabaseConnection{
						CockroachDB: &CockroachDBConnection{
							URI: ValueOrValueFrom{
								ValueFrom: &ValueFrom{
									Vault: &Vault{
										Secret: "database/creds/test",
									},
								},
							},
						},
					},
				},
			},
			want: `
{{- with secret "database/creds/test" -}}
postgres://{{ .Data.username }}:{{ .Data.password }}@postgres:5432/testdb{{- end }}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test := test
			t.Parallel()

			a, err := test.db.GetVaultAnnotations()

			if err != nil {
				t.Fatal(err)
			}

			if got := a["vault.hashicorp.com/agent-inject-template-schemaherouri"]; got != test.want {
				t.Fatalf("Expected:\n%s\ngot:\n%s", test.want, got)
			}
		})
	}
}
