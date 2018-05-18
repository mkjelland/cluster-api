
/*
Copyright 2018 The Kubernetes Authors.

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


package cluster

import (
	"log"

	"github.com/kubernetes-incubator/apiserver-builder/pkg/builders"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/typed/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/controller/sharedinformers"
	listers "sigs.k8s.io/cluster-api/pkg/client/listers_generated/cluster/v1alpha1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"github.com/golang/glog"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/cluster-api/util"
)

// +controller:group=cluster,version=v1alpha1,kind=Cluster,resource=clusters
type ClusterControllerImpl struct {
	builders.DefaultControllerFns

	actuator Actuator
	// lister indexes properties about Cluster
	lister listers.ClusterLister

	kubernetesClientSet *kubernetes.Clientset
	clientSet           clientset.Interface
	clusterClient       v1alpha1.ClusterInterface
}

// Init initializes the controller and is called by the generated code
// Register watches for additional resource types here.
func (c *ClusterControllerImpl) Init(arguments sharedinformers.ControllerInitArguments, actuator Actuator) {
	// Use the lister for indexing clusters labels
	c.lister = arguments.GetSharedInformers().Factory.Cluster().V1alpha1().Clusters().Lister()
	c.actuator = actuator

	cs, err := clientset.NewForConfig(arguments.GetRestConfig())
	if err != nil {
		glog.Fatalf("error creating cluster client: %v", err)
	}
	c.clientSet = cs
	c.kubernetesClientSet = arguments.GetSharedInformers().KubernetesClientSet

	// Create cluster actuator.
	// TODO: Assume default namespace for now. Maybe a separate a controller per namespace?
	c.clusterClient = cs.ClusterV1alpha1().Clusters(corev1.NamespaceDefault)
	c.actuator = actuator
}

// Reconcile handles enqueued messages. The delete will be handled by finalizer.
func (c *ClusterControllerImpl) Reconcile(cluster *clusterv1.Cluster) error {
	name := cluster.Name
	log.Printf("Running reconcile Cluster for %s\n", name)

	if !cluster.ObjectMeta.DeletionTimestamp.IsZero() {
		// no-op if finalizer has been removed.
		if !util.Contains(cluster.ObjectMeta.Finalizers, clusterv1.ClusterFinalizer) {
			glog.Infof("reconciling cluster object %v causes a no-op as there is no finalizer.", name)
			return nil
		}

		glog.Infof("reconciling cluster object %v triggers delete.", name)
		if err := c.actuator.Delete(cluster); err != nil {
			glog.Errorf("Error deleting cluster object %v; %v", name, err)
			return err
		}
		// Remove finalizer on successful deletion.
		glog.Infof("cluster object %v deletion successful, removing finalizer.", name)
		cluster.ObjectMeta.Finalizers = util.Filter(cluster.ObjectMeta.Finalizers, clusterv1.ClusterFinalizer)
		if _, err := c.clusterClient.Update(cluster); err != nil {
			glog.Errorf("Error removing finalizer from cluster object %v; %v", name, err)
			return err
		}
		return nil
	}

	exist, err := c.actuator.Exists(cluster)
	if err != nil {
		glog.Errorf("Error checking existance of cluster instance for cluster object %v; %v", name, err)
		return err
	}
	if exist {
		glog.Infof("reconciling cluster object %v triggers idempotent update.", name)
		return c.actuator.Update(cluster)
	}
	// Cluster resource created. Cluster does not yet exist.
	glog.Infof("reconciling cluster object %v triggers idempotent create.", cluster.ObjectMeta.Name)
	return c.actuator.Create(cluster, nil)
}

func (c *ClusterControllerImpl) Get(namespace, name string) (*clusterv1.Cluster, error) {
	return c.lister.Clusters(namespace).Get(name)
}