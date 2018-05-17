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
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// Actuator controls clusters on a specific infrastructure. All
// methods should be idempotent unless otherwise specified.
type Actuator interface {
	// Create the cluster.
	Create(*clusterv1.Cluster, []*clusterv1.Machine) error
	// Delete the cluster.
	Delete(*clusterv1.Cluster) error
	// Update the cluster to the provided definition.
	Update(*clusterv1.Cluster) error
	// Checks if the cluster currently exists.
	Exists(*clusterv1.Cluster) (bool, error)
}