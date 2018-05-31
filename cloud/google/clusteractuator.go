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
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"sigs.k8s.io/cluster-api/cloud/google/clients"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	client "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/typed/cluster/v1alpha1"
	gceconfigv1 "sigs.k8s.io/cluster-api/cloud/google/gceproviderconfig/v1alpha1"
)

type GCEClusterClient struct {
	computeService GCEClientComputeService
	clusterClient  client.ClusterInterface
	gceProviderConfigCodec   *gceconfigv1.GCEProviderConfigCodec
}

type ClusterActuatorParams struct {
	ComputeService GCEClientComputeService
	ClusterClient  client.ClusterInterface
}

func NewClusterActuator(params ClusterActuatorParams) (*GCEClusterClient, error) {
	computeService, err := getOrNewComputeServiceForCluster(params)
	if err != nil {
		return nil, err
	}

	codec, err := gceconfigv1.NewCodec()
	if err != nil {
		return nil, err
	}

	return &GCEClusterClient{
		computeService: computeService,
		clusterClient:  params.ClusterClient,
		gceProviderConfigCodec: codec,
	}, nil
}

func (gce *GCEClusterClient) Reconcile(cluster *clusterv1.Cluster) error {
	clusterConfig, err := ClusterProviderConfig(cluster.Spec.ProviderConfig, gce.gceProviderConfigCodec)
	if err != nil {
		return err
	}
	if GetMasterServiceAccount(cluster) == "" {
		err = CreateMasterNodeServiceAccount(cluster, clusterConfig.Project)
		if err != nil {
			return err
		}
	}
	if GetWorkerServiceAccount(cluster) == "" {
		err = CreateWorkerNodeServiceAccount(cluster, clusterConfig.Project)
		if err != nil {
			return err
		}
	}
	return nil
}

func (gce *GCEClusterClient) Delete(cluster *clusterv1.Cluster) error {
	clusterConfig, err := ClusterProviderConfig(cluster.Spec.ProviderConfig, gce.gceProviderConfigCodec)
	if err != nil {
		return fmt.Errorf("Cannot unmarshal cluster's providerConfig field: %v", err)
	}
	if err := DeleteMasterNodeServiceAccount(cluster, clusterConfig.Project); err != nil {
		return fmt.Errorf("error deleting master node service account: %v", err)
	}
	if err := DeleteWorkerNodeServiceAccount(cluster, clusterConfig.Project); err != nil {
		return fmt.Errorf("error deleting worker node service account: %v", err)
	}
	return nil
}

func getOrNewComputeServiceForCluster(params ClusterActuatorParams) (GCEClientComputeService, error) {
	if params.ComputeService != nil {
		return params.ComputeService, nil
	}
	// The default GCP client expects the environment variable
	// GOOGLE_APPLICATION_CREDENTIALS to point to a file with service credentials.
	client, err := google.DefaultClient(context.TODO(), compute.ComputeScope)
	if err != nil {
		return nil, err
	}
	computeService, err := clients.NewComputeService(client)
	if err != nil {
		return nil, err
	}
	return computeService, nil
}

func ClusterProviderConfig(providerConfig clusterv1.ProviderConfig, codec *gceconfigv1.GCEProviderConfigCodec,) (*gceconfigv1.GCEClusterProviderConfig, error) {
	var config gceconfigv1.GCEClusterProviderConfig
	err := codec.DecodeFromProviderConfig(providerConfig, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
