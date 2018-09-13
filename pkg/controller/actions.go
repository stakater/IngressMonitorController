package controller

import (
	"log"

	"github.com/stakater/IngressMonitorController/pkg/constants"
	"github.com/stakater/IngressMonitorController/pkg/kube"
)

// Action is an interface for ingress and route actions
type Action interface {
	handle(c *MonitorController) error
	getNames(c *MonitorController) (string, string)
}

// ResourceUpdatedAction provide implementation of action interface
type ResourceUpdatedAction struct {
	resource    interface{}
	oldResource interface{}
}

// ResourceDeletedAction provide implementation of action interface
type ResourceDeletedAction struct {
	resource interface{}
}

func (r ResourceUpdatedAction) getNames(c *MonitorController) (string, string) {
	rAFuncs := kube.GetResourceActionFuncs(r.resource)
	monitorName := c.getMonitorName(rAFuncs, r.resource)
	if r.oldResource == nil {
		return monitorName, monitorName
	}

	oldMonitorName := c.getMonitorName(rAFuncs, r.oldResource)
	return monitorName, oldMonitorName
}

func (r ResourceUpdatedAction) handle(c *MonitorController) error {
	rAFuncs := kube.GetResourceActionFuncs(r.resource)

	monitorName, oldMonitorName := r.getNames(c)
	monitorURL := c.getMonitorURL(r.resource)

	log.Println("Monitor Name: " + monitorName)
	log.Println("Monitor URL: " + monitorURL)

	annotations := rAFuncs.AnnotationFunc(r.resource)
	if value, ok := annotations[constants.MonitorEnabledAnnotation]; ok {
		if value == "true" {
			// Annotation exists and is enabled
			c.createOrUpdateMonitors(monitorName, oldMonitorName, monitorURL, annotations)
		} else {
			// Annotation exists but is disabled
			c.removeMonitorsIfExist(oldMonitorName)
		}

	} else {
		c.removeMonitorsIfExist(oldMonitorName)
		log.Println("Not doing anything with this ingress because no annotation exists with name: " + constants.MonitorEnabledAnnotation)
	}

	return nil
}

func (r ResourceDeletedAction) getNames(c *MonitorController) (string, string) {
	rAFuncs := kube.GetResourceActionFuncs(r.resource)
	monitorName := c.getMonitorName(rAFuncs, r.resource)
	return monitorName, monitorName
}

func (r ResourceDeletedAction) handle(c *MonitorController) error {
	if c.config.EnableMonitorDeletion {
		// Delete the monitor if it exists
		monitorName, _ := r.getNames(c)
		c.removeMonitorsIfExist(monitorName)
	}

	return nil
}
