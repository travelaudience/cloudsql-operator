module github.com/travelaudience/cloudsql-postgres-operator

go 1.12

replace k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d

replace k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible

replace k8s.io/kubernetes => k8s.io/kubernetes v1.14.1

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/appscode/jsonpatch v0.0.0-20190108182946-7c0e3b262f30
	github.com/evanphx/json-patch v4.2.0+incompatible // indirect
	github.com/glendc/go-external-ip v0.0.0-20170425150139-139229dcdddd
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/lib/pq v1.1.1
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/pflag v1.0.3 // indirect
	golang.org/x/crypto v0.0.0-20190513172903-22d7a77e9e5f // indirect
	golang.org/x/net v0.0.0-20190509222800-a4d6f7feada5 // indirect
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	google.golang.org/api v0.5.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0-20190512063542-eae0ddcf85ba
	k8s.io/apiextensions-apiserver v0.0.0-20190514064203-3f96d5001990
	k8s.io/apimachinery v0.0.0-20190514012558-1f207b29b441
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/kube-openapi v0.0.0-20190510232812-a01b7d5d6c22 // indirect
	k8s.io/kubernetes v1.14.1
	k8s.io/utils v0.0.0-20190506122338-8fab8cb257d5 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)
