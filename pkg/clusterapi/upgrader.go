package clusterapi

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
)

type Upgrader struct {
	*clients
}

func NewUpgrader(capiClient CAPIClient, kubectlClient KubectlClient) *Upgrader {
	return &Upgrader{
		clients: &clients{
			capiClient:    capiClient,
			kubectlClient: kubectlClient,
		},
	}
}

func (u *Upgrader) Upgrade(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, currentSpec, newSpec *cluster.Spec) (*types.ChangeDiff, error) {
	logger.V(1).Info("Checking for CAPI upgrades")
	if !newSpec.Cluster.IsSelfManaged() {
		logger.V(1).Info("Skipping CAPI upgrades, not a self-managed cluster")
		return nil, nil
	}

	capiChangeDiff := capiChangeDiff(currentSpec, newSpec, provider)
	if capiChangeDiff == nil {
		logger.V(1).Info("Nothing to upgrade for CAPI")
		return nil, nil
	}

	logger.V(1).Info("Starting CAPI upgrades")
	if err := u.capiClient.Upgrade(ctx, managementCluster, provider, newSpec, capiChangeDiff); err != nil {
		return nil, fmt.Errorf("failed upgrading ClusterAPI from bundles %d to bundles %d: %v", currentSpec.Bundles.Spec.Number, newSpec.Bundles.Spec.Number, err)
	}

	return capiChangeDiff.toChangeDiff(), nil
}

type CAPIChangeDiff struct {
	CertManager            *types.ComponentChangeDiff
	Core                   *types.ComponentChangeDiff
	ControlPlane           *types.ComponentChangeDiff
	BootstrapProviders     []types.ComponentChangeDiff
	InfrastructureProvider *types.ComponentChangeDiff
}

func (c *CAPIChangeDiff) toChangeDiff() *types.ChangeDiff {
	if c == nil {
		logger.V(1).Info("Nothing to upgrade for CAPI")
		return nil
	}
	r := make([]*types.ComponentChangeDiff, 0, 4+len(c.BootstrapProviders))
	r = append(r, c.CertManager, c.Core, c.ControlPlane, c.InfrastructureProvider)
	for _, bootstrapChangeDiff := range c.BootstrapProviders {
		b := bootstrapChangeDiff
		r = append(r, &b)
	}

	return types.NewChangeDiff(r...)
}

func CapiChangeDiff(currentSpec, newSpec *cluster.Spec, provider providers.Provider) *types.ChangeDiff {
	return capiChangeDiff(currentSpec, newSpec, provider).toChangeDiff()
}

func capiChangeDiff(currentSpec, newSpec *cluster.Spec, provider providers.Provider) *CAPIChangeDiff {
	changeDiff := &CAPIChangeDiff{}
	componentChanged := false

	cvb, err := currentSpec.GetCPVersionsBundle()
	if err != nil {
		return nil
	}
	nvb, err := newSpec.GetCPVersionsBundle()
	if err != nil {
		return nil
	}

	if cvb.CertManager.Version != nvb.CertManager.Version {
		changeDiff.CertManager = &types.ComponentChangeDiff{
			ComponentName: "cert-manager",
			NewVersion:    nvb.CertManager.Version,
			OldVersion:    cvb.CertManager.Version,
		}
		logger.V(1).Info("Cert-manager change diff", "oldVersion", changeDiff.CertManager.OldVersion, "newVersion", changeDiff.CertManager.NewVersion)
		componentChanged = true
	}

	if cvb.ClusterAPI.Version != nvb.ClusterAPI.Version {
		changeDiff.Core = &types.ComponentChangeDiff{
			ComponentName: "cluster-api",
			NewVersion:    nvb.ClusterAPI.Version,
			OldVersion:    cvb.ClusterAPI.Version,
		}
		logger.V(1).Info("CAPI Core change diff", "oldVersion", changeDiff.Core.OldVersion, "newVersion", changeDiff.Core.NewVersion)
		componentChanged = true
	}

	if cvb.ControlPlane.Version != nvb.ControlPlane.Version {
		changeDiff.ControlPlane = &types.ComponentChangeDiff{
			ComponentName: "kubeadm",
			NewVersion:    nvb.ControlPlane.Version,
			OldVersion:    cvb.ControlPlane.Version,
		}
		logger.V(1).Info("CAPI Control Plane provider change diff", "oldVersion", changeDiff.ControlPlane.OldVersion, "newVersion", changeDiff.ControlPlane.NewVersion)
		componentChanged = true
	}

	if cvb.Bootstrap.Version != nvb.Bootstrap.Version {
		componentChangeDiff := types.ComponentChangeDiff{
			ComponentName: "kubeadm",
			NewVersion:    nvb.Bootstrap.Version,
			OldVersion:    cvb.Bootstrap.Version,
		}
		changeDiff.BootstrapProviders = append(changeDiff.BootstrapProviders, componentChangeDiff)
		logger.V(1).Info("CAPI Kubeadm Bootstrap Provider change diff", "oldVersion", componentChangeDiff.OldVersion, "newVersion", componentChangeDiff.NewVersion)
		componentChanged = true
	}

	if cvb.ExternalEtcdBootstrap.Version != nvb.ExternalEtcdBootstrap.Version {
		componentChangeDiff := types.ComponentChangeDiff{
			ComponentName: "etcdadm-bootstrap",
			NewVersion:    nvb.ExternalEtcdBootstrap.Version,
			OldVersion:    cvb.ExternalEtcdBootstrap.Version,
		}
		changeDiff.BootstrapProviders = append(changeDiff.BootstrapProviders, componentChangeDiff)
		logger.V(1).Info("CAPI Etcdadm Bootstrap Provider change diff", "oldVersion", componentChangeDiff.OldVersion, "newVersion", componentChangeDiff.NewVersion)
		componentChanged = true
	}

	if cvb.ExternalEtcdController.Version != nvb.ExternalEtcdController.Version {
		componentChangeDiff := types.ComponentChangeDiff{
			ComponentName: "etcdadm-controller",
			NewVersion:    nvb.ExternalEtcdController.Version,
			OldVersion:    cvb.ExternalEtcdController.Version,
		}
		changeDiff.BootstrapProviders = append(changeDiff.BootstrapProviders, componentChangeDiff)
		logger.V(1).Info("CAPI Etcdadm Controller Provider change diff", "oldVersion", componentChangeDiff.OldVersion, "newVersion", componentChangeDiff.NewVersion)
		componentChanged = true
	}

	if providerChangeDiff := provider.ChangeDiff(currentSpec, newSpec); providerChangeDiff != nil {
		changeDiff.InfrastructureProvider = providerChangeDiff
		logger.V(1).Info("CAPI Infrastrcture Provider change diff", "provider", providerChangeDiff.ComponentName, "oldVersion", providerChangeDiff.OldVersion, "newVersion", providerChangeDiff.NewVersion)
		componentChanged = true
	}

	if !componentChanged {
		return nil
	}

	return changeDiff
}
