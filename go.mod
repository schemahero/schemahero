module github.com/schemahero/schemahero

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.11.0
	github.com/aws/aws-sdk-go-v2/config v1.9.0
	github.com/aws/aws-sdk-go-v2/credentials v1.5.0
	github.com/aws/aws-sdk-go-v2/service/ssm v1.14.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gocql/gocql v0.0.0-20200815110948-5378c8f664e9
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/jackc/pgx/v4 v4.13.0
	github.com/mattn/go-sqlite3 v1.14.9
	github.com/onsi/gomega v1.16.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	github.com/xo/dburl v0.0.0-20200124232849-e9ec94f52bc3
	go.uber.org/zap v1.19.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.3
	k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery v0.22.3
	k8s.io/cli-runtime v0.22.3
	k8s.io/client-go v0.22.3
	k8s.io/kube-openapi v0.0.0-20210527164424-3c818078ee3d // indirect
	sigs.k8s.io/controller-runtime v0.10.2
)

replace github.com/appscode/jsonpatch => github.com/gomodules/jsonpatch v2.0.1+incompatible
