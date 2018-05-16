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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	gceconfigv1 "sigs.k8s.io/cluster-api/cloud/google/gceproviderconfig/v1alpha1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type GCEClusterClient struct {
	scheme                   *runtime.Scheme
	codecFactory             *serializer.CodecFactory
}

func NewClusterActuator() (*GCEClusterClient, error) {
	scheme, codecFactory, err := gceconfigv1.NewSchemeAndCodecs()
	if err != nil {
		return nil, err
	}

	return &GCEClusterClient{
		scheme: scheme,
		codecFactory:   codecFactory,
	}, nil
}

func (gce *GCEClusterClient) Create(cluster *clusterv1.Cluster, initialMachines []*clusterv1.Machine) error {
	if err := CreateMasterNodeServiceAccount(cluster); err != nil {
		return fmt.Errorf("error creating master node service account: %v", err)
	}
	if err := CreateWorkerNodeServiceAccount(cluster); err != nil {
		return fmt.Errorf("error creating worker node service account: %v", err)
	}
	return nil
}

func (gce *GCEClusterClient) Delete(cluster *clusterv1.Cluster) error {
	if err := DeleteMasterNodeServiceAccount(cluster); err != nil {
		return fmt.Errorf("error deleting master node service account: %v", err)
	}
	if err := DeleteWorkerNodeServiceAccount(cluster); err != nil {
		return fmt.Errorf("error deleting worker node service account: %v", err)
	}
	return nil
}

func (gce *GCEClusterClient) Update(cluster *clusterv1.Cluster) error {
	return fmt.Errorf("NYI: Cluster Updates are not yet supported")
}

func (gce *GCEClusterClient) Exists(cluster *clusterv1.Cluster) (bool, error) {
	return false, fmt.Errorf("NYI: Cluster Exists method is not yet implemented")
}


