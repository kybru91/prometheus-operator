// Copyright The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers"
	"k8s.io/client-go/tools/cache"
)

// AlertmanagerConfigLister helps list AlertmanagerConfigs.
// All objects returned here must be treated as read-only.
type AlertmanagerConfigLister interface {
	// List lists all AlertmanagerConfigs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.AlertmanagerConfig, err error)
	// AlertmanagerConfigs returns an object that can list and get AlertmanagerConfigs.
	AlertmanagerConfigs(namespace string) AlertmanagerConfigNamespaceLister
	AlertmanagerConfigListerExpansion
}

// alertmanagerConfigLister implements the AlertmanagerConfigLister interface.
type alertmanagerConfigLister struct {
	listers.ResourceIndexer[*v1alpha1.AlertmanagerConfig]
}

// NewAlertmanagerConfigLister returns a new AlertmanagerConfigLister.
func NewAlertmanagerConfigLister(indexer cache.Indexer) AlertmanagerConfigLister {
	return &alertmanagerConfigLister{listers.New[*v1alpha1.AlertmanagerConfig](indexer, v1alpha1.Resource("alertmanagerconfig"))}
}

// AlertmanagerConfigs returns an object that can list and get AlertmanagerConfigs.
func (s *alertmanagerConfigLister) AlertmanagerConfigs(namespace string) AlertmanagerConfigNamespaceLister {
	return alertmanagerConfigNamespaceLister{listers.NewNamespaced[*v1alpha1.AlertmanagerConfig](s.ResourceIndexer, namespace)}
}

// AlertmanagerConfigNamespaceLister helps list and get AlertmanagerConfigs.
// All objects returned here must be treated as read-only.
type AlertmanagerConfigNamespaceLister interface {
	// List lists all AlertmanagerConfigs in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.AlertmanagerConfig, err error)
	// Get retrieves the AlertmanagerConfig from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.AlertmanagerConfig, error)
	AlertmanagerConfigNamespaceListerExpansion
}

// alertmanagerConfigNamespaceLister implements the AlertmanagerConfigNamespaceLister
// interface.
type alertmanagerConfigNamespaceLister struct {
	listers.ResourceIndexer[*v1alpha1.AlertmanagerConfig]
}
