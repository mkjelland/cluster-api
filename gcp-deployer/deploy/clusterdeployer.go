package deploy

import (
	"sigs.k8s.io/cluster-api/pkg/controller/cluster"
)

// Provider-specific machine logic the deployer needs.
type clusterDeployer interface {
	cluster.Actuator
}
