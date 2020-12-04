module github.com/stakater/IngressMonitorController

go 1.15

require (
	cloud.google.com/go v0.49.0
	github.com/Azure/azure-sdk-for-go v44.0.0+incompatible
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.0
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.0 // indirect
	github.com/antoineaugusti/updown v0.0.0-20190412074625-d590ab97f115
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/openshift/api v0.0.0-20200526144822-34f54f12813a
	github.com/operator-framework/operator-sdk v0.19.0
	github.com/rs/zerolog v1.20.0
	github.com/russellcardullo/go-pingdom v1.3.0
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.5.1
	google.golang.org/api v0.14.0
	google.golang.org/genproto v0.0.0-20200117163144-32f20d992d24
	gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.1
)

replace k8s.io/client-go => k8s.io/client-go v0.18.2
