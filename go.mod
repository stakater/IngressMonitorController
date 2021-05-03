module github.com/stakater/IngressMonitorController

go 1.15

require (
	cloud.google.com/go v0.54.0
	github.com/Azure/azure-sdk-for-go v44.0.0+incompatible
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.0
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.0 // indirect
	github.com/antoineaugusti/updown v0.0.0-20190412074625-d590ab97f115
	github.com/go-logr/logr v0.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/onsi/ginkgo v1.15.0 // indirect
	github.com/onsi/gomega v1.10.5 // indirect
	github.com/openshift/api v0.0.0-20200526144822-34f54f12813a
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/russellcardullo/go-pingdom v1.3.0
	github.com/stakater/operator-utils v0.1.13
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20201012173705-84dcc777aaee // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/tools v0.1.0 // indirect
	google.golang.org/api v0.20.0
	google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a
	google.golang.org/grpc v1.30.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.2 // indirect
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)

replace github.com/stakater/IngressMonitorController => ./github.com/stakater/IngressMonitorController

// replace k8s.io/client-go => k8s.io/client-go v0.18.2
