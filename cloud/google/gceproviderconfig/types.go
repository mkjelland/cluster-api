/*
Copyright 2017 The Kubernetes Authors.

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

package gceproviderconfig

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GCEMachineProviderConfig struct {
	metav1.TypeMeta `json:",inline"`

	Zone        string `json:"zone"`
	MachineType string `json:"machineType"`

	// The name of the OS to be installed on the machine.
	OS    string `json:"os"`
	Disks []Disk `json:"disks"`
}

type Disk struct {
	InitializeParams DiskInitializeParams `json:"initializeParams"`
}

type DiskInitializeParams struct {
	DiskSizeGb int64  `json:"diskSizeGb"`
	DiskType   string `json:"diskType"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GCEClusterProviderConfig struct {
	metav1.TypeMeta `json:",inline"`

	Project     string `json:"project"`
}