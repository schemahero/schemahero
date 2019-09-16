module github.com/schemahero/schemahero

go 1.12

require (
	github.com/go-sql-driver/mysql v1.4.1
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/lib/pq v1.1.1
	github.com/onsi/gomega v1.5.0
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.3.0
	github.com/weaveworks/gitops-toolkit v0.0.0-20190830163251-b6682e98e2fa
	github.com/xo/dburl v0.0.0-20190203050942-98997a05b24f
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/net v0.0.0-20190912160710-24e19bdeb0f2
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0-beta.4
	sigs.k8s.io/controller-tools v0.2.0-beta.2 // indirect
)

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d

replace github.com/appscode/jsonpatch => github.com/gomodules/jsonpatch v2.0.0+incompatible
