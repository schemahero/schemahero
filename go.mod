module github.com/schemahero/schemahero

go 1.14

require (
	github.com/aws/aws-sdk-go-v2 v0.23.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gocql/gocql v0.0.0-20200815110948-5378c8f664e9
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/jackc/pgx/v4 v4.10.1
	github.com/jmespath/go-jmespath v0.3.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.6
	github.com/onsi/gomega v1.10.5
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/xo/dburl v0.0.0-20200124232849-e9ec94f52bc3
	go.uber.org/zap v1.16.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/cli-runtime v0.18.0
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.5.1-0.20200402191424-df180accb901
)

replace github.com/appscode/jsonpatch => github.com/gomodules/jsonpatch v2.0.1+incompatible
