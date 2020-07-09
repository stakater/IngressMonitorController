/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	ingressmonitorv1alpha1 "github.com/stakater/IngressMonitorController/pkg/apis/ingressmonitor/v1alpha1"
	versioned "github.com/stakater/IngressMonitorController/pkg/client/clientset/versioned"
	internalinterfaces "github.com/stakater/IngressMonitorController/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/stakater/IngressMonitorController/pkg/client/listers/ingressmonitor/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// IngressMonitorInformer provides access to a shared informer and lister for
// IngressMonitors.
type IngressMonitorInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.IngressMonitorLister
}

type ingressMonitorInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewIngressMonitorInformer constructs a new informer for IngressMonitor type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewIngressMonitorInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredIngressMonitorInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredIngressMonitorInformer constructs a new informer for IngressMonitor type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredIngressMonitorInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.IngressmonitorV1alpha1().IngressMonitors(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.IngressmonitorV1alpha1().IngressMonitors(namespace).Watch(context.TODO(), options)
			},
		},
		&ingressmonitorv1alpha1.IngressMonitor{},
		resyncPeriod,
		indexers,
	)
}

func (f *ingressMonitorInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredIngressMonitorInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *ingressMonitorInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&ingressmonitorv1alpha1.IngressMonitor{}, f.defaultInformer)
}

func (f *ingressMonitorInformer) Lister() v1alpha1.IngressMonitorLister {
	return v1alpha1.NewIngressMonitorLister(f.Informer().GetIndexer())
}
