package v1alpha4

import (
	"testing"

	"github.com/stretchr/testify/require"
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
										AgentInject: true,
										Role:        "test",
										Secret:      "database/creds/test",
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
										AgentInject: true,
										Role:        "test",
										Secret:      "database/creds/test",
									},
								},
							},
						},
					},
				},
			},
			want: `
{{- with secret "database/creds/test" -}}
{{ .Data.username }}:{{ .Data.password }}@tcp(mysql:3306)/testdb{{- end }}`,
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
										AgentInject: true,
										Role:        "test",
										Secret:      "database/creds/test",
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
			name: "Postgres_template_passed_in",
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
										AgentInject:        true,
										Role:               "test",
										Secret:             "database/creds/test",
										ConnectionTemplate: "postgres://{{ .username }}:{{ .password }}@postgres:1234/userdb",
									},
								},
							},
						},
					},
				},
			},
			want: `
{{- with secret "database/creds/test" -}}
postgres://{{ .username }}:{{ .password }}@postgres:1234/userdb{{- end }}`,
		},
		{
			name: "rqlite",
			db: &Database{
				ObjectMeta: v1.ObjectMeta{
					Name: "testdb",
				},
				Spec: DatabaseSpec{
					Connection: DatabaseConnection{
						RQLite: &RqliteConnection{
							URI: ValueOrValueFrom{
								ValueFrom: &ValueFrom{
									Vault: &Vault{
										AgentInject: true,
										Role:        "test",
										Secret:      "database/creds/test",
									},
								},
							},
						},
					},
				},
			},
			want: `
{{- with secret "database/creds/test" -}}
http://{{ .Data.username }}:{{ .Data.password }}@rqlite:4001/{{- end }}`,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req := require.New(t)

			a, err := test.db.GetVaultAnnotations()
			req.NoError(err)

			if got := a["vault.hashicorp.com/agent-inject-template-schemaherouri"]; got != test.want {
				t.Fatalf("Expected:\n%s\ngot:\n%s", test.want, got)
			}
		})
	}
}
