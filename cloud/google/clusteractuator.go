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

package google

import (
	"fmt"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type GCEClusterClient struct {
}

func NewClusterActuator() (*GCEClusterClient, error) {
	return &GCEClusterClient{
	}, nil
}

func (gce *GCEClusterClient) Create(cluster *clusterv1.Cluster, initialMachines []*clusterv1.Machine) error {
	return fmt.Errorf("NYI: Cluster Creations are not yet supported")
}

func (gce *GCEClusterClient) Delete(cluster *clusterv1.Cluster) error {
	return fmt.Errorf("NYI: Cluster Deletions are not yet supported")
}

func (gce *GCEClusterClient) Update(cluster *clusterv1.Cluster) error {
	return fmt.Errorf("NYI: Cluster Updates are not yet supported")
}

func (gce *GCEClusterClient) Exists(cluster *clusterv1.Cluster) (bool, error) {
	return false, fmt.Errorf("NYI: Cluster Exists method is not yet implemented")
}
