package cluster

import (
	"fmt"
	"strings"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"

	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type Spec struct {
	*Config
	Bundles           *v1alpha1.Bundles
	VersionsBundle    *VersionsBundle
	eksdRelease       *eksdv1alpha1.Release
	OIDCConfig        *eksav1alpha1.OIDCConfig
	AWSIamConfig      *eksav1alpha1.AWSIamConfig
	ManagementCluster *types.Cluster // TODO(g-gaston): cleanup, this doesn't belong here
	WorkerVersions    map[string]*WorkerVersions
}

// WorkerVersions holds the VersionsBundle and eksdRelease for node groups that contain a KubernetesVersion.
type WorkerVersions struct {
	VersionsBundle *VersionsBundle
	eksdRelease    *eksdv1alpha1.Release
}

func (w *WorkerVersions) deepCopy() *WorkerVersions {
	return &WorkerVersions{
		VersionsBundle: &VersionsBundle{
			VersionsBundle: w.VersionsBundle.VersionsBundle.DeepCopy(),
			KubeDistro:     w.VersionsBundle.KubeDistro.deepCopy(),
		},
		eksdRelease: w.eksdRelease.DeepCopy(),
	}
}

func deepCopyWorkerVersions(workerVersions map[string]*WorkerVersions) map[string]*WorkerVersions {
	if workerVersions == nil {
		return nil
	}
	retVal := map[string]*WorkerVersions{}
	for key, wv := range workerVersions {
		retVal[key] = wv.deepCopy()
	}
	return retVal
}

func (s *Spec) DeepCopy() *Spec {
	return &Spec{
		Config:       s.Config.DeepCopy(),
		OIDCConfig:   s.OIDCConfig.DeepCopy(),
		AWSIamConfig: s.AWSIamConfig.DeepCopy(),
		VersionsBundle: &VersionsBundle{
			VersionsBundle: s.VersionsBundle.VersionsBundle.DeepCopy(),
			KubeDistro:     s.VersionsBundle.KubeDistro.deepCopy(),
		},
		eksdRelease:    s.eksdRelease.DeepCopy(),
		Bundles:        s.Bundles.DeepCopy(),
		WorkerVersions: deepCopyWorkerVersions(s.WorkerVersions),
	}
}

type VersionsBundle struct {
	*v1alpha1.VersionsBundle
	KubeDistro *KubeDistro
}

// EKSD represents an eks-d release.
type EKSD struct {
	// Channel is the minor Kubernetes version for the eks-d release (eg. "1.23", "1.24", etc.)
	Channel string
	// Number is the monotonically increasing number that distinguishes the different eks-d releases
	// for the same Kubernetes minor version (channel).
	Number int
}

type KubeDistro struct {
	EKSD                EKSD
	Kubernetes          VersionedRepository
	CoreDNS             VersionedRepository
	Etcd                VersionedRepository
	NodeDriverRegistrar v1alpha1.Image
	LivenessProbe       v1alpha1.Image
	ExternalAttacher    v1alpha1.Image
	ExternalProvisioner v1alpha1.Image
	Pause               v1alpha1.Image
	EtcdImage           v1alpha1.Image
	EtcdVersion         string
	AwsIamAuthImage     v1alpha1.Image
	KubeProxy           v1alpha1.Image
}

func (k *KubeDistro) deepCopy() *KubeDistro {
	k2 := *k
	return &k2
}

type VersionedRepository struct {
	Repository, Tag string
}

// NewSpec builds a new [Spec].
func NewSpec(config *Config, bundles *v1alpha1.Bundles, eksdRelease *eksdv1alpha1.Release, workerEksdReleases map[string]*eksdv1alpha1.Release) (*Spec, error) {
	s := &Spec{}

	versionsBundle, err := constructVersionsBundles(config.Cluster.Spec.KubernetesVersion, bundles, eksdRelease)
	if err != nil {
		return nil, err
	}

	workerVersions := map[string]*WorkerVersions{}
	for _, wng := range config.Cluster.Spec.WorkerNodeGroupConfigurations {
		if weksd, ok := workerEksdReleases[wng.Name]; ok && wng.KubernetesVersion != nil {
			vb, err := constructVersionsBundles(*wng.KubernetesVersion, bundles, weksd)
			if err != nil {
				return nil, err
			}
			wv := &WorkerVersions{VersionsBundle: vb, eksdRelease: weksd}
			workerVersions[wng.Name] = wv
		}
	}

	s.Bundles = bundles
	s.Config = config
	s.VersionsBundle = versionsBundle
	s.eksdRelease = eksdRelease
	s.WorkerVersions = workerVersions

	// Get first aws iam config if it exists
	// Config supports multiple configs because Cluster references a slice
	// But we validate that only one of each type is referenced
	for _, ac := range s.Config.AWSIAMConfigs {
		s.AWSIamConfig = ac
		break
	}

	// Get first oidc config if it exists
	for _, oc := range s.Config.OIDCConfigs {
		s.OIDCConfig = oc
		break
	}

	return s, nil
}

func constructVersionsBundles(version eksav1alpha1.KubernetesVersion, bundles *v1alpha1.Bundles, eksd *eksdv1alpha1.Release) (*VersionsBundle, error) {
	vb, err := GetVersionsBundle(version, bundles)
	if err != nil {
		return nil, err
	}

	kubeDistro, err := buildKubeDistro(eksd)
	if err != nil {
		return nil, err
	}

	return &VersionsBundle{
		VersionsBundle: vb,
		KubeDistro:     kubeDistro,
	}, nil
}

func (s *Spec) KubeDistroImages() []v1alpha1.Image {
	images := []v1alpha1.Image{}
	for _, component := range s.eksdRelease.Status.Components {
		for _, asset := range component.Assets {
			if asset.Image != nil {
				images = append(images, v1alpha1.Image{URI: asset.Image.URI})
			}
		}
	}
	return images
}

func buildKubeDistro(eksd *eksdv1alpha1.Release) (*KubeDistro, error) {
	kubeDistro := &KubeDistro{
		EKSD: EKSD{
			Channel: eksd.Spec.Channel,
			Number:  eksd.Spec.Number,
		},
	}
	assets := make(map[string]*eksdv1alpha1.AssetImage)
	for _, component := range eksd.Status.Components {
		for _, asset := range component.Assets {
			if asset.Image != nil {
				assets[asset.Name] = asset.Image
			}
		}
		if component.Name == "etcd" {
			kubeDistro.EtcdVersion = strings.TrimPrefix(component.GitTag, "v")
		}
	}

	kubeDistroComponents := map[string]*v1alpha1.Image{
		"node-driver-registrar-image": &kubeDistro.NodeDriverRegistrar,
		"livenessprobe-image":         &kubeDistro.LivenessProbe,
		"external-attacher-image":     &kubeDistro.ExternalAttacher,
		"external-provisioner-image":  &kubeDistro.ExternalProvisioner,
		"pause-image":                 &kubeDistro.Pause,
		"etcd-image":                  &kubeDistro.EtcdImage,
		"aws-iam-authenticator-image": &kubeDistro.AwsIamAuthImage,
		"kube-proxy-image":            &kubeDistro.KubeProxy,
	}

	for assetName, image := range kubeDistroComponents {
		i := assets[assetName]
		if i == nil {
			return nil, fmt.Errorf("asset %s is no present in eksd release %s", assetName, eksd.Spec.Channel)
		}

		image.URI = i.URI
	}

	kubeDistroRepositories := map[string]*VersionedRepository{
		"coredns-image":        &kubeDistro.CoreDNS,
		"etcd-image":           &kubeDistro.Etcd,
		"kube-apiserver-image": &kubeDistro.Kubernetes,
	}

	for assetName, image := range kubeDistroRepositories {
		i := assets[assetName]
		if i == nil {
			return nil, fmt.Errorf("asset %s is not present in eksd release %s", assetName, eksd.Spec.Channel)
		}

		image.Repository, image.Tag = kubeDistroRepository(i)
	}

	return kubeDistro, nil
}

func kubeDistroRepository(image *eksdv1alpha1.AssetImage) (repo, tag string) {
	i := v1alpha1.Image{
		URI: image.URI,
	}

	lastInd := strings.LastIndex(i.Image(), "/")
	if lastInd == -1 {
		return i.Image(), i.Tag()
	}

	return i.Image()[:lastInd], i.Tag()
}

func (vb *VersionsBundle) KubeDistroImages() []v1alpha1.Image {
	var images []v1alpha1.Image
	images = append(images, vb.KubeDistro.EtcdImage)
	images = append(images, vb.KubeDistro.ExternalAttacher)
	images = append(images, vb.KubeDistro.ExternalProvisioner)
	images = append(images, vb.KubeDistro.LivenessProbe)
	images = append(images, vb.KubeDistro.NodeDriverRegistrar)
	images = append(images, vb.KubeDistro.Pause)

	return images
}

func (vb *VersionsBundle) Images() []v1alpha1.Image {
	var images []v1alpha1.Image
	images = append(images, vb.VersionsBundle.Images()...)
	images = append(images, vb.KubeDistroImages()...)

	return images
}

func (vb *VersionsBundle) Ovas() []v1alpha1.Archive {
	return vb.VersionsBundle.Ovas()
}

func BundlesRefDefaulter() Defaulter {
	return func(c *Config) error {
		if c.Cluster.Spec.BundlesRef == nil {
			c.Cluster.Spec.BundlesRef = &eksav1alpha1.BundlesRef{}
		}
		return nil
	}
}
