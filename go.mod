module github.com/schemahero/schemahero

go 1.12

require (
	github.com/Azure/go-autorest v11.1.2+incompatible // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/lib/pq v1.1.1
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.8.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/teris-io/shortid v0.0.0-20171029131806-771a37caa5cf
	github.com/ventu-io/go-shortid v0.0.0-20171029131806-771a37caa5cf
	github.com/xo/dburl v0.0.0-20190203050942-98997a05b24f
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	gopkg.in/square/go-jose.v2 v2.3.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.18.0
	k8s.io/apiextensions-apiserver v0.18.0
	k8s.io/apimachinery v0.18.0
	k8s.io/cli-runtime v0.17.0
	k8s.io/client-go v0.18.0
	k8s.io/code-generator v0.18.1-beta.0 // indirect
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/controller-tools v0.2.8 // indirect
)

replace github.com/appscode/jsonpatch => github.com/gomodules/jsonpatch v2.0.1+incompatible
