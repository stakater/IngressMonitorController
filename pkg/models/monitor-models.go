package models

import (
	ingressmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/ingressmonitor/v1alpha1"
	"github.com/stakater/IngressMonitorController/pkg/kube"
)

type Monitor struct {
	URL         string
	Name        string
	ID          string
	Annotations map[string]string
}

func NewMonitor(monitorName string, monitorSpec ingressmonitorv1alpha1.IngressMonitorSpec) (Monitor) {
	monitorUrl := kube.getURL(monitorSpec)
	return Monitor{
		Name: monitorName,
		URL:  monitorUrl,
	}
}
