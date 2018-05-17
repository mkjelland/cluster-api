package google_test

import (
	compute "google.golang.org/api/compute/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/cloud/google"
	gceconfigv1 "sigs.k8s.io/cluster-api/cloud/google/gceproviderconfig/v1alpha1"
	"sigs.k8s.io/cluster-api/cloud/google/machinesetup"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"strings"
	"testing"
)

type GCEClientComputeServiceMock struct {
	mockImagesGet           func(project string, image string) (*compute.Image, error)
	mockImagesGetFromFamily func(project string, family string) (*compute.Image, error)
	mockInstancesDelete     func(project string, zone string, targetInstance string) (*compute.Operation, error)
	mockInstancesGet        func(project string, zone string, instance string) (*compute.Instance, error)
	mockInstancesInsert     func(project string, zone string, instance *compute.Instance) (*compute.Operation, error)
	mockZoneOperationsGet   func(project string, zone string, operation string) (*compute.Operation, error)
}

func (c *GCEClientComputeServiceMock) ImagesGet(project string, image string) (*compute.Image, error) {
	if c.mockImagesGet == nil {
		return nil, nil
	}
	return c.mockImagesGet(project, image)
}

func (c *GCEClientComputeServiceMock) ImagesGetFromFamily(project string, family string) (*compute.Image, error) {
	if c.mockImagesGetFromFamily == nil {
		return nil, nil
	}
	return c.mockImagesGetFromFamily(project, family)
}

func (c *GCEClientComputeServiceMock) InstancesDelete(project string, zone string, targetInstance string) (*compute.Operation, error) {
	if c.mockInstancesDelete == nil {
		return nil, nil
	}
	return c.mockInstancesDelete(project, zone, targetInstance)
}

func (c *GCEClientComputeServiceMock) InstancesGet(project string, zone string, instance string) (*compute.Instance, error) {
	if c.mockInstancesGet == nil {
		return nil, nil
	}
	return c.mockInstancesGet(project, zone, instance)
}

func (c *GCEClientComputeServiceMock) InstancesInsert(project string, zone string, instance *compute.Instance) (*compute.Operation, error) {
	if c.mockInstancesInsert == nil {
		return nil, nil
	}
	return c.mockInstancesInsert(project, zone, instance)
}

func (c *GCEClientComputeServiceMock) ZoneOperationsGet(project string, zone string, operation string) (*compute.Operation, error) {
	if c.mockZoneOperationsGet == nil {
		return nil, nil
	}
	return c.mockZoneOperationsGet(project, zone, operation)
}

type GCEClientMachineSetupConfigMock struct {
	mockGetYaml     func() (string, error)
	mockGetImage    func(params *machinesetup.ConfigParams) (string, error)
	mockGetMetadata func(params *machinesetup.ConfigParams) (machinesetup.Metadata, error)
}

func (m *GCEClientMachineSetupConfigMock) GetYaml() (string, error) {
	if m.mockGetYaml == nil {
		return "", nil
	}
	return m.mockGetYaml()
}

func (m *GCEClientMachineSetupConfigMock) GetImage(params *machinesetup.ConfigParams) (string, error) {
	if m.mockGetYaml == nil {
		return "", nil
	}
	return m.mockGetImage(params)
}

func (m *GCEClientMachineSetupConfigMock) GetMetadata(params *machinesetup.ConfigParams) (machinesetup.Metadata, error) {
	if m.mockGetYaml == nil {
		return machinesetup.Metadata{}, nil
	}
	return m.mockGetMetadata(params)
}

func TestNoDisks(t *testing.T) {
	config := newGCEMachineProviderConfigFixture()
	config.Disks = make([]gceconfigv1.Disk, 0)
	receivedInstance, computeServiceMock := newInsertInstanceCapturingMock()
	createCluster(t, config, computeServiceMock)
	checkInstanceValues(t, receivedInstance, 0)
}

func TestMinimumSizeShouldBeEnforced(t *testing.T) {
	config := newGCEMachineProviderConfigFixture()
	config.Disks = []gceconfigv1.Disk{
		{
			InitializeParams: gceconfigv1.DiskInitializeParams{
				DiskType:   "pd-ssd",
				DiskSizeGb: int64(6),
			},
		},
	}
	receivedInstance, computeServiceMock := newInsertInstanceCapturingMock()
	createCluster(t, config, computeServiceMock)
	checkInstanceValues(t, receivedInstance, 1)
	checkDiskValues(t, receivedInstance.Disks[0], true, 30, "pd-ssd", "projects/ubuntu-os-cloud/global/images/family/ubuntu-1710")
}

func TestOneDisk(t *testing.T) {
	config := newGCEMachineProviderConfigFixture()
	config.Disks = []gceconfigv1.Disk{
		{
			InitializeParams: gceconfigv1.DiskInitializeParams{
				DiskType:   "pd-ssd",
				DiskSizeGb: 37,
			},
		},
	}
	receivedInstance, computeServiceMock := newInsertInstanceCapturingMock()
	createCluster(t, config, computeServiceMock)
	checkInstanceValues(t, receivedInstance, 1)
	checkDiskValues(t, receivedInstance.Disks[0], true, 37, "pd-ssd", "projects/ubuntu-os-cloud/global/images/family/ubuntu-1710")
}

func TestTwoDisks(t *testing.T) {
	config := newGCEMachineProviderConfigFixture()
	config.Disks = []gceconfigv1.Disk{
		{
			InitializeParams: gceconfigv1.DiskInitializeParams{
				DiskType:   "pd-ssd",
				DiskSizeGb: 32,
			},
		},
		{
			InitializeParams: gceconfigv1.DiskInitializeParams{
				DiskType:   "pd-standard",
				DiskSizeGb: 45,
			},
		},
	}
	receivedInstance, computeServiceMock := newInsertInstanceCapturingMock()
	createCluster(t, config, computeServiceMock)
	checkInstanceValues(t, receivedInstance, 2)
	checkDiskValues(t, receivedInstance.Disks[0], true, 32, "pd-ssd", "projects/ubuntu-os-cloud/global/images/family/ubuntu-1710")
	checkDiskValues(t, receivedInstance.Disks[1], false, 45, "pd-standard", "")
}

func checkInstanceValues(t *testing.T, instance *compute.Instance, diskCount int) {
	t.Helper()
	if instance == nil {
		t.Error("expected a valid instance")
	}
	if len(instance.Disks) != diskCount {
		t.Errorf("invalid disk count: expected '%v' got '%v'", diskCount, len(instance.Disks))
	}
}

func checkDiskValues(t *testing.T, disk *compute.AttachedDisk, boot bool, sizeGb int64, diskType string, image string) {
	t.Helper()
	if disk.Boot != boot {
		t.Errorf("invalid disk.Boot value: expected '%v' got '%v'", boot, disk.Boot)
	}
	if disk.InitializeParams.DiskSizeGb != sizeGb {
		t.Errorf("invalid disk size: expected '%v' got '%v'", sizeGb, disk.InitializeParams.DiskSizeGb)
	}
	if !strings.Contains(disk.InitializeParams.DiskType, diskType) {
		t.Errorf("invalid disk type '%v': expected it to contain '%v'", disk.InitializeParams.DiskType, diskType)
	}
	if disk.InitializeParams.SourceImage != image {
		t.Errorf("invalid image: expected '%v' got '%v'", image, disk.InitializeParams.SourceImage)
	}
}

func createCluster(t *testing.T, config gceconfigv1.GCEMachineProviderConfig, computeServiceMock *GCEClientComputeServiceMock) {
	cluster := newDefaultClusterFixture(t)
	machine := newMachine(t, config)
	configWatch := newMachineSetupConfigWatcher()
	params := google.MachineActuatorParams{
		ComputeService:           computeServiceMock,
		MachineSetupConfigGetter: configWatch,
	}
	gce, err := google.NewMachineActuator(params)
	if err != nil {
		t.Fatalf("unable to create machine actuator: %v", err)
	}
	err = gce.Create(cluster, machine)
	if err != nil {
		t.Fatalf("unable to create cluster: %v", err)
	}
}

func newInsertInstanceCapturingMock() (*compute.Instance, *GCEClientComputeServiceMock) {
	var receivedInstance compute.Instance
	computeServiceMock := GCEClientComputeServiceMock{
		mockInstancesInsert: func(project string, zone string, instance *compute.Instance) (*compute.Operation, error) {
			receivedInstance = *instance
			return &compute.Operation{
				Status: "DONE",
			}, nil
		},
	}
	return &receivedInstance, &computeServiceMock
}

func newMachineSetupConfigMock() *GCEClientMachineSetupConfigMock {
	return &GCEClientMachineSetupConfigMock{
		mockGetYaml: func() (string, error) {
			return "", nil
		},
		mockGetMetadata: func(params *machinesetup.ConfigParams) (machinesetup.Metadata, error) {
			metadata := machinesetup.Metadata{}
			return metadata, nil
		},
		mockGetImage: func(params *machinesetup.ConfigParams) (string, error) {
			return "image-name", nil
		},
	}
}

type TestMachineSetupConfigWatcher struct {
	machineSetupConfigMock *GCEClientMachineSetupConfigMock
}

func newMachineSetupConfigWatcher() *TestMachineSetupConfigWatcher {
	return &TestMachineSetupConfigWatcher{
		machineSetupConfigMock: newMachineSetupConfigMock(),
	}
}

func (cw *TestMachineSetupConfigWatcher) GetMachineSetupConfig() (machinesetup.MachineSetupConfig, error) {
	return cw.machineSetupConfigMock, nil
}

func newMachine(t *testing.T, gceProviderConfig gceconfigv1.GCEMachineProviderConfig) *v1alpha1.Machine {
	gceProviderConfigCodec, err := gceconfigv1.NewCodec()
	if err != nil {
		t.Fatalf("unable to create GCE provider config codec: %v", err)
	}
	var providerConfig v1alpha1.ProviderConfig
	err = gceProviderConfigCodec.EncodeToProviderConfig(&gceProviderConfig, &providerConfig)
	if err != nil {
		t.Fatalf("unable to encode provider config: %v", err)
	}

	return &v1alpha1.Machine{
		Spec: v1alpha1.MachineSpec{
			ProviderConfig: providerConfig,
			Versions: v1alpha1.MachineVersionInfo{
				Kubelet:      "1.9.4",
				ControlPlane: "1.9.4",
				ContainerRuntime: v1alpha1.ContainerRuntimeInfo{
					Name:    "docker",
					Version: "1.12.0",
				},
			},
			Roles: []common.MachineRole{
				common.MasterRole,
			},
		},
	}
}

func newGCEMachineProviderConfigFixture() gceconfigv1.GCEMachineProviderConfig {
	return gceconfigv1.GCEMachineProviderConfig{
		TypeMeta: v1.TypeMeta{
			APIVersion: "gceproviderconfig/v1alpha1",
			Kind:       "GCEMachineProviderConfig",
		},
		Zone:    "us-west5-f",
		OS:      "os-name",
		Disks:   make([]gceconfigv1.Disk, 0),
	}
}

func newGCEClusterProviderConfigFixture() gceconfigv1.GCEClusterProviderConfig {
	return gceconfigv1.GCEClusterProviderConfig{
		TypeMeta: v1.TypeMeta{
			APIVersion: "gceproviderconfig/v1alpha1",
			Kind:       "GCEClusterProviderConfig",
		},
		Project:    "project-name-2000",
	}
}

func newDefaultClusterFixture(t *testing.T) *v1alpha1.Cluster {
	gceProviderConfig := newGCEClusterProviderConfigFixture()
	gceProviderConfigCodec, err := gceconfigv1.NewCodec()
	if err != nil {
		t.Fatalf("unable to create GCE provider config codec: %v", err)
	}
	var providerConfig v1alpha1.ProviderConfig
	err = gceProviderConfigCodec.EncodeToProviderConfig(&gceProviderConfig, &providerConfig)
	if err != nil {
		t.Fatalf("unable to encode provider config: %v", err)
	}

	return &v1alpha1.Cluster{
		TypeMeta: v1.TypeMeta{
			Kind: "Cluster",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "cluster-test",
		},
		Spec: v1alpha1.ClusterSpec{
			ClusterNetwork: v1alpha1.ClusterNetworkingConfig{
				Services: v1alpha1.NetworkRanges{
					CIDRBlocks: []string{
						"10.96.0.0/12",
					},
				},
				Pods: v1alpha1.NetworkRanges{
					CIDRBlocks: []string{
						"192.168.0.0/16",
					},
				},
			},
			ProviderConfig: providerConfig,
		},
	}
}
