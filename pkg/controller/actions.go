package controller

import (
	"log"

	"github.com/stakater/IngressMonitorController/pkg/kube/wrappers"
	"k8s.io/api/extensions/v1beta1"
)

type IngressAction interface {
	handle(c *MonitorController) error
	getNames(c *MonitorController) (string, string)
}

type IngressUpdatedAction struct {
	ingress    *v1beta1.Ingress
	oldIngress *v1beta1.Ingress
}

type IngressDeletedAction struct {
	ingress *v1beta1.Ingress
}

func (i IngressUpdatedAction) getNames(c *MonitorController) (string, string) {
	monitorName := c.getMonitorName(i.ingress)
	if i.oldIngress == nil {
		return monitorName, monitorName
	}

	oldMonitorName := c.getMonitorName(i.oldIngress)
	return monitorName, oldMonitorName
}

func (i IngressUpdatedAction) handle(c *MonitorController) error {
	monitorName, oldMonitorName := i.getNames(c)
	monitorURL := c.getMonitorURL(i.ingress)

	log.Println("Monitor Name: " + monitorName)
	log.Println("Monitor URL: " + monitorURL)

	annotations := i.ingress.GetAnnotations()
	if value, ok := annotations[wrappers.MonitorEnabledAnnotation]; ok {
		if value == "true" {
			// Annotation exists and is enabled
			c.createOrUpdateMonitors(monitorName, oldMonitorName, monitorURL, annotations)
		} else {
			// Annotation exists but is disabled
			c.removeMonitorsIfExist(oldMonitorName)
		}

	} else {
		c.removeMonitorsIfExist(oldMonitorName)
		log.Println("Not doing anything with this ingress because no annotation exists with name: " + wrappers.MonitorEnabledAnnotation)
	}

	return nil
}

func (i IngressDeletedAction) getNames(c *MonitorController) (string, string) {
	monitorName := c.getMonitorName(i.ingress)
	return monitorName, monitorName
}

func (i IngressDeletedAction) handle(c *MonitorController) error {
	if c.config.EnableMonitorDeletion {
		// Delete the monitor if it exists
		monitorName, _ := i.getNames(c)
		c.removeMonitorsIfExist(monitorName)
	}

	return nil
}
