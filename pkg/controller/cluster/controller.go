
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

	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/controller/sharedinformers"
	listers "sigs.k8s.io/cluster-api/pkg/client/listers_generated/cluster/v1alpha1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"github.com/golang/glog"
)

// +controller:group=cluster,version=v1alpha1,kind=Cluster,resource=clusters
type ClusterControllerImpl struct {
	builders.DefaultControllerFns

	actuator Actuator
	// lister indexes properties about Cluster
	lister listers.ClusterLister
}

// Init initializes the controller and is called by the generated code
// Register watches for additional resource types here.
func (c *ClusterControllerImpl) Init(arguments sharedinformers.ControllerInitArguments, actuator Actuator) {
	// Use the lister for indexing clusters labels
	c.lister = arguments.GetSharedInformers().Factory.Cluster().V1alpha1().Clusters().Lister()
	c.actuator = actuator
}

// Reconcile handles enqueued messages
func (c *ClusterControllerImpl) Reconcile(cluster *v1alpha1.Cluster) error {
	// Implement controller logic here
	name := cluster.Name
	log.Printf("Running reconcile Cluster for %s\n", name)

	exist, err := c.actuator.Exists(cluster)
	if err != nil {
		glog.Errorf("Error checking existance of cluster instance for cluster object %v; %v", name, err)
		return err
	}
	if exist {
		glog.Infof("reconciling machine object %v triggers idempotent update.", name)
		return c.actuator.Update(cluster)
	}
	// Machine resource created. Machine does not yet exist.
	glog.Infof("reconciling machine object %v triggers idempotent create.", cluster.ObjectMeta.Name)
	return c.create(cluster, nil)
	return nil
}

func (c *ClusterControllerImpl) create(cluster *clusterv1.Cluster, machines []*clusterv1.Machine) error {
	return c.actuator.Create(cluster, machines)
}


func (c *ClusterControllerImpl) Get(namespace, name string) (*v1alpha1.Cluster, error) {
	return c.lister.Clusters(namespace).Get(name)
}
