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

package main

import (
	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/util/logs"
	machineoptions "sigs.k8s.io/cluster-api/cloud/google/cmd/gce-controller/machine-controller-app/options"
	"sigs.k8s.io/cluster-api/cloud/google/cmd/gce-controller/machine-controller-app"
	clusteroptions "sigs.k8s.io/cluster-api/cloud/google/cmd/gce-controller/cluster-controller-app/options"
	"sigs.k8s.io/cluster-api/cloud/google/cmd/gce-controller/cluster-controller-app"
	"sigs.k8s.io/cluster-api/pkg/controller/config"
)

func main() {

	fs := pflag.CommandLine
	var controllerType string
	fs.StringVar(&controllerType, "controller", controllerType, "machine or cluster controller")

	machineServer := machineoptions.NewMachineControllerServer()
	machineServer.AddFlags(fs)

	clusterServer := clusteroptions.NewClusterControllerServer()
	clusterServer.AddFlags(fs)

	config.ControllerConfig.AddFlags(pflag.CommandLine)

	pflag.Parse()

	logs.InitLogs()
	defer logs.FlushLogs()

	if controllerType == "machine" {
		if err := machine_controller_app.RunMachineController(machineServer); err != nil {
			glog.Errorf("Failed to start machine controller. Err: %v", err)
		}
	} else if controllerType == "cluster" {
		if err := cluster_controller_app.RunClusterController(clusterServer); err != nil {
			glog.Errorf("Failed to start machine controller. Err: %v", err)
		}
	}

	glog.Errorf("Failed to start controller, `controller` flag must be either `machine` or `cluster` but was %v.", controllerType)
}
