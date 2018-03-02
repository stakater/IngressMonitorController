package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const monitorEnabledAnnotation = "monitor.stakater.com/enabled"

// MonitorController which can be used for monitoring ingresses
type MonitorController struct {
	clientset       *kubernetes.Clientset
	namespace       string
	indexer         cache.Indexer
	queue           workqueue.RateLimitingInterface
	informer        cache.Controller
	monitorServices []MonitorServiceProxy
	config          Config
}

func NewMonitorController(namespace string, clientset *kubernetes.Clientset, config Config) *MonitorController {
	controller := &MonitorController{
		clientset: clientset,
		namespace: namespace,
		config:    config,
	}

	controller.monitorServices = setupMonitorServicesForProviders(config.Providers)

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Create the Ingress Watcher
	ingressListWatcher := cache.NewListWatchFromClient(clientset.ExtensionsV1beta1().RESTClient(), "ingresses", namespace, fields.Everything())

	indexer, informer := cache.NewIndexerInformer(ingressListWatcher, &v1beta1.Ingress{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.onIngressAdded,
		UpdateFunc: controller.onIngressUpdated,
		DeleteFunc: controller.onIngressDeleted,
	}, cache.Indexers{})
	controller.indexer = indexer
	controller.informer = informer
	controller.queue = queue

	return controller
}

func setupMonitorServicesForProviders(providers []Provider) []MonitorServiceProxy {
	if len(providers) < 1 {
		log.Panic("Cannot Instantiate controller with no providers")
	}

	monitorServices := []MonitorServiceProxy{}

	for index := 0; index < len(providers); index++ {
		monitorServices = append(monitorServices, providers[index].createMonitorService())
	}

	return monitorServices
}

func (c *MonitorController) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	log.Println("Starting Ingress Monitor controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	log.Println("Stopping Ingress Monitor controller")
}

func (c *MonitorController) runWorker() {
	for c.processNextItem() {
	}
}

func (c *MonitorController) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two ingresses with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.handleIngress(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

// handleIngress handles sync between the provided monitors for each ingress
func (c *MonitorController) handleIngress(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		log.Printf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		c.handleIngressOnDeletion(key)

	} else {
		ingress := obj.(*v1beta1.Ingress)
		c.handleIngressOnCreationOrUpdation(ingress)
	}
	return nil
}

func (c *MonitorController) handleIngressOnDeletion(key string) {
	if c.config.EnableMonitorDeletion {
		// Delete the monitor if it exists
		// key is in the format "namespace/ingressname"
		splitted := strings.Split(key, "/")
		monitorName := c.getMonitorName(splitted[1], c.namespace)

		c.removeMonitorsIfExist(monitorName)
	}
}

func (c *MonitorController) getMonitorName(ingressName string, namespace string) string {
	format, err := getNameTemplateFormat(c.config.MonitorNameTemplate)
	if err != nil {
		log.Fatal("Failed to parse MonitorNameTemplate")
	}
	return fmt.Sprintf(format, ingressName, namespace)
}

func (c *MonitorController) getMonitorURL(ingress *v1beta1.Ingress) string {
	var url string
	if ingress.Spec.TLS != nil && len(ingress.Spec.TLS) > 0 {
		// Use https for TLS
		url = "https://" + ingress.Spec.TLS[0].Hosts[0]
	} else {
		url = "http://" + ingress.Spec.Rules[0].Host
	}

	url += c.getIngressSubPathWithPort(ingress.Spec.Rules)

	return url
}

func (c *MonitorController) getIngressSubPathWithPort(rules []v1beta1.IngressRule) string {
	if rules != nil && len(rules) > 0 {
		port := c.getIngressPort(rules[0])
		subPath := c.getIngressSubPath(rules[0])
		return port + subPath
	}
	return ""
}

func (c *MonitorController) getIngressPort(rule v1beta1.IngressRule) string {
	if rule.HTTP.Paths != nil && len(rule.HTTP.Paths) > 0 {
		return rule.HTTP.Paths[0].Backend.ServicePort.StrVal
	}
	return ""
}

func (c *MonitorController) getIngressSubPath(rule v1beta1.IngressRule) string {
	if rule.HTTP.Paths != nil && len(rule.HTTP.Paths) > 0 {
		return rule.HTTP.Paths[0].Path
	}
	return ""
}

func (c *MonitorController) handleIngressOnCreationOrUpdation(ingress *v1beta1.Ingress) {
	monitorName := c.getMonitorName(ingress.GetName(), c.namespace)
	monitorURL := c.getMonitorURL(ingress)

	log.Println("Monitor Name: " + monitorName)
	log.Println("Monitor URL: " + monitorURL)

	annotations := ingress.GetAnnotations()

	if value, ok := annotations[monitorEnabledAnnotation]; ok {
		if value == "true" {
			// Annotation exists and is enabled
			c.createOrUpdateMonitors(monitorName, monitorURL)
		} else {
			// Annotation exists but is disabled
			c.removeMonitorsIfExist(monitorName)
		}

	} else {
		c.removeMonitorsIfExist(monitorName)
		log.Println("Not doing anything with this ingress because no annotation exists with name: " + monitorEnabledAnnotation)
	}
}

func (c *MonitorController) removeMonitorsIfExist(monitorName string) {
	for index := 0; index < len(c.monitorServices); index++ {
		c.removeMonitorIfExists(c.monitorServices[index], monitorName)
	}
}

func (c *MonitorController) removeMonitorIfExists(monitorService MonitorServiceProxy, monitorName string) {
	m, _ := monitorService.GetByName(monitorName)

	if m != nil { // Monitor Exists
		monitorService.Remove(*m) // Remove the monitor
	} else {
		log.Println("Cannot find monitor with name: " + monitorName)
	}
}

func (c *MonitorController) createOrUpdateMonitors(monitorName string, monitorURL string) {
	for index := 0; index < len(c.monitorServices); index++ {
		monitorService := c.monitorServices[index]
		c.createOrUpdateMonitor(monitorService, monitorName, monitorURL)
	}
}

func (c *MonitorController) createOrUpdateMonitor(monitorService MonitorServiceProxy, monitorName string, monitorURL string) {
	m, _ := monitorService.GetByName(monitorName)

	if m != nil { // Monitor Already Exists
		log.Println("Monitor already exists for ingress: " + monitorName)
		if m.url != monitorURL { // Monitor does not have the same url
			// update the monitor with the new url
			m.url = monitorURL
			monitorService.Update(*m)
		}
	} else {
		// Create a new monitor for this ingress
		m := Monitor{name: monitorName, url: monitorURL}
		monitorService.Add(m)
	}
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *MonitorController) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		log.Printf("Error syncing ingress %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	log.Printf("Dropping ingress %q out of the queue: %v", key, err)
}

func (c *MonitorController) onIngressAdded(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err == nil {
		c.queue.Add(key)
	}
}

func (c *MonitorController) onIngressUpdated(old interface{}, new interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(new)
	if err == nil {
		c.queue.Add(key)
	}
}

func (c *MonitorController) onIngressDeleted(obj interface{}) {
	// IndexerInformer uses a delta queue, therefore for deletes we have to use this
	// key function.
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err == nil {
		c.queue.Add(key)
	} else {
		log.Println("Error: " + err.Error())
	}
}
