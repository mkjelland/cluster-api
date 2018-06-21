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
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"sigs.k8s.io/cluster-api/cloud/google/clients"
	gceconfigv1 "sigs.k8s.io/cluster-api/cloud/google/gceproviderconfig/v1alpha1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	client "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/typed/cluster/v1alpha1"
	"k8s.io/client-go/tools/record"
	apierrors "sigs.k8s.io/cluster-api/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

const (
	firewallRuleAnnotationPrefix = "gce.clusterapi.k8s.io/firewall"
	firewallRuleInternalSuffix   = "-allow-cluster-internal"
	firewallRuleApiSuffix        = "-allow-api-public"
)

type GCEClusterClient struct {
	computeService         GCEClientComputeService
	clusterClient          client.ClusterInterface
	gceProviderConfigCodec *gceconfigv1.GCEProviderConfigCodec
	eventRecorder  record.EventRecorder
}

type ClusterActuatorParams struct {
	ComputeService GCEClientComputeService
	ClusterClient  client.ClusterInterface
	EventRecorder  record.EventRecorder
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
		computeService:         computeService,
		clusterClient:          params.ClusterClient,
		gceProviderConfigCodec: codec,
		eventRecorder:  params.EventRecorder,
	}, nil
}

func (gce *GCEClusterClient) Reconcile(cluster *clusterv1.Cluster) error {
	glog.Infof("Reconciling cluster %v.", cluster.Name)
	existed, err := gce.createFirewallRuleIfNotExists(cluster, &compute.Firewall{
		Name:    cluster.Name + firewallRuleInternalSuffix,
		Network: "global/networks/default",
		Allowed: []*compute.FirewallAllowed{
			{
				IPProtocol: "tcp",
			},
		},
		TargetTags: []string{cluster.Name + "-worker"},
		SourceTags: []string{cluster.Name + "-worker"},
	})
	if err != nil {
		glog.Warningf("Error creating firewall rule for internal cluster traffic: %v", err)
	}
	existed, err = gce.createFirewallRuleIfNotExists(cluster, &compute.Firewall{
		Name:    cluster.Name + firewallRuleApiSuffix,
		Network: "global/networks/default",
		Allowed: []*compute.FirewallAllowed{
			{
				IPProtocol: "tcp",
				Ports:      []string{"443"},
			},
		},
		TargetTags:   []string{"https-server"},
		SourceRanges: []string{"0.0.0.0/0"},
	})
	if err != nil {
		glog.Warningf("Error creating firewall rule for core api server traffic: %v", err)
	}

	if !existed {

	}

	return nil
}

func (gce *GCEClusterClient) Delete(cluster *clusterv1.Cluster) error {
	err := gce.deleteFirewallRule(cluster, cluster.Name+firewallRuleInternalSuffix)
	if err != nil {
		return fmt.Errorf("error deleting firewall rule for internal cluster traffic: %v", err)
	}
	err = gce.deleteFirewallRule(cluster, cluster.Name+firewallRuleApiSuffix)
	if err != nil {
		return fmt.Errorf("error deleting firewall rule for core api server traffic: %v", err)
	}

	gce.eventRecorder.Eventf(cluster, corev1.EventTypeNormal, "Deleted", "Deleted Cluster %v", cluster.Name)

	return nil
}

func getOrNewComputeServiceForCluster(params ClusterActuatorParams) (GCEClientComputeService, error) {
	if params.ComputeService != nil {
		return params.ComputeService, nil
	}
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

func (gce *GCEClusterClient) createFirewallRuleIfNotExists(cluster *clusterv1.Cluster, firewallRule *compute.Firewall) (bool, error) {
	ruleExists, ok := cluster.ObjectMeta.Annotations[firewallRuleAnnotationPrefix+firewallRule.Name]
	if ok && ruleExists == "true" {
		// The firewall rule was already created.
		return true, nil
	}
	clusterConfig, err := gce.clusterproviderconfig(cluster.Spec.ProviderConfig)
	if err != nil {
		return false, fmt.Errorf("error parsing cluster provider config: %v", err)
	}
	firewallRules, err := gce.computeService.FirewallsGet(clusterConfig.Project)
	if err != nil {
		return false, fmt.Errorf("error getting firewall rules: %v", err)
	}

	if !gce.containsFirewallRule(firewallRules, firewallRule.Name) {
		op, err := gce.computeService.FirewallsInsert(clusterConfig.Project, firewallRule)
		if err != nil {
			return false, fmt.Errorf("error creating firewall rule: %v", err)
		}
		err = gce.computeService.WaitForOperation(clusterConfig.Project, op)
		if err != nil {
			return false, fmt.Errorf("error waiting for firewall rule creation: %v", err)
		}
	}
	// TODO (mkjelland) move this to a GCEClusterProviderStatus #347
	if cluster.ObjectMeta.Annotations == nil {
		cluster.ObjectMeta.Annotations = make(map[string]string)
	}
	cluster.ObjectMeta.Annotations[firewallRuleAnnotationPrefix+firewallRule.Name] = "true"
	_, err = gce.clusterClient.Update(cluster)
	if err != nil {
		fmt.Errorf("error updating cluster annotations %v", err)
	}
	return false, nil
}

func (gce *GCEClusterClient) containsFirewallRule(firewallRules *compute.FirewallList, ruleName string) bool {
	for _, rule := range firewallRules.Items {
		if ruleName == rule.Name {
			return true
		}
	}
	return false
}

func (gce *GCEClusterClient) deleteFirewallRule(cluster *clusterv1.Cluster, ruleName string) error {
	clusterConfig, err := gce.clusterproviderconfig(cluster.Spec.ProviderConfig)
	if err != nil {
		return fmt.Errorf("error parsing cluster provider config: %v", err)
	}
	op, err := gce.computeService.FirewallsDelete(clusterConfig.Project, ruleName)
	if err != nil {
		return fmt.Errorf("error deleting firewall rule: %v", err)
	}
	return gce.computeService.WaitForOperation(clusterConfig.Project, op)
}

func (gce *GCEClusterClient) clusterproviderconfig(providerConfig clusterv1.ProviderConfig) (*gceconfigv1.GCEClusterProviderConfig, error) {
	var config gceconfigv1.GCEClusterProviderConfig
	err := gce.gceProviderConfigCodec.DecodeFromProviderConfig(providerConfig, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (gce *GCEClusterClient) handleClusterError(cluster *clusterv1.Cluster, err *apierrors.ClusterError, eventAction string) error {
	if gce.clusterClient != nil {
		reason := err.Reason
		message := err.Message
		cluster.Status.ErrorReason = &reason
		cluster.Status.ErrorMessage = &message
		gce.clusterClient.Update(cluster)
		if eventAction != "" {
			gce.eventRecorder.Eventf(cluster, corev1.EventTypeWarning, "Failed"+eventAction, "%v", err.Reason)
		}
	}

	glog.Errorf("Machine error: %v", err.Message)
	return err
}