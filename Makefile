VERSION := 1.0.14
HELMVALUES := kubernetes/chart/ingressmonitorcontroller/values.yaml
HELMNAME := ingressmonitorcontroller

list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | xargs

clean:
	rm -rvf out

build:
	go build -o out/ingressmonitorcontroller github.com/BoseCorp/IngressMonitorController/src

docker-build:
	docker run --rm -it -v .

helm-template:
	helm template kubernetes/chart/ingressmonitorcontroller --values $(HELMVALUES) --name $(HELMNAME)

helm-install:
	helm install kubernetes/chart/ingressmonitorcontroller --values $(HELMVALUES) --name $(HELMNAME)