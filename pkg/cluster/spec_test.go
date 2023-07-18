package cluster_test

import (
	"embed"
	"testing"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/manifests/eksd"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

//go:embed testdata
var testdataFS embed.FS

func TestNewSpecError(t *testing.T) {
	tests := []struct {
		name        string
		config      *cluster.Config
		bundles     *releasev1.Bundles
		eksdRelease map[anywherev1.KubernetesVersion]*eksdv1.Release
		error       string
	}{
		{
			name: "no VersionsBundle for kube version",
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					Spec: anywherev1.ClusterSpec{
						KubernetesVersion: anywherev1.Kube124,
					},
				},
			},
			bundles: &releasev1.Bundles{
				Spec: releasev1.BundlesSpec{
					Number: 2,
				},
			},
			eksdRelease: map[anywherev1.KubernetesVersion]*eksdv1.Release{
				anywherev1.Kube119: {},
			},
			error: "kubernetes version 1.24 is not supported by bundles manifest 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(cluster.NewSpec(tt.config, tt.bundles, tt.eksdRelease)).Error().To(
				MatchError(ContainSubstring(tt.error)),
			)
		})
	}
}

func TestNewSpecValid(t *testing.T) {
	g := NewWithT(t)
	kube124 := anywherev1.KubernetesVersion("1.24")
	config := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			Spec: anywherev1.ClusterSpec{
				KubernetesVersion: kube124,
				WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
					{
						KubernetesVersion: &kube124,
					},
				},
			},
		},
		OIDCConfigs: map[string]*anywherev1.OIDCConfig{
			"myconfig": {},
		},
		AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{
			"myconfig": {},
		},
	}
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			Number: 2,
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.24",
				},
			},
		},
	}
	eksd := test.EksdReleasesMap()

	spec, err := cluster.NewSpec(config, bundles, eksd)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(spec.AWSIamConfig).NotTo(BeNil())
	g.Expect(spec.OIDCConfig).NotTo(BeNil())
}

func TestSpecDeepCopy(t *testing.T) {
	g := NewWithT(t)
	r := files.NewReader()
	yaml, err := r.ReadFile("testdata/docker_cluster_oidc_awsiam_flux.yaml")
	g.Expect(err).To(Succeed())
	config, err := cluster.ParseConfig(yaml)
	g.Expect(err).To(Succeed())
	bundles := test.Bundles(t)
	eksd := test.EksdReleasesMap()
	spec, err := cluster.NewSpec(config, bundles, eksd)
	g.Expect(err).To(Succeed())

	g.Expect(spec.DeepCopy()).To(Equal(spec))
}

func TestBundlesRefDefaulter(t *testing.T) {
	tests := []struct {
		name         string
		bundles      *releasev1.Bundles
		config, want *cluster.Config
	}{
		{
			name: "no bundles ref",
			bundles: &releasev1.Bundles{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bundles-1",
					Namespace: "eksa-system",
				},
			},
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{},
			},
			want: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					Spec: anywherev1.ClusterSpec{
						BundlesRef: &anywherev1.BundlesRef{},
					},
				},
			},
		},
		{
			name: "with previous bundles ref",
			bundles: &releasev1.Bundles{
				ObjectMeta: metav1.ObjectMeta{
					Name: "bundles-1",
				},
			},
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					Spec: anywherev1.ClusterSpec{
						BundlesRef: &anywherev1.BundlesRef{
							Name:       "bundles-2",
							Namespace:  "default",
							APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
						},
					},
				},
			},
			want: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					Spec: anywherev1.ClusterSpec{
						BundlesRef: &anywherev1.BundlesRef{
							Name:       "bundles-2",
							Namespace:  "default",
							APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			defaulter := cluster.BundlesRefDefaulter()
			g.Expect(defaulter(tt.config)).To(Succeed())
			g.Expect(tt.config).To(Equal(tt.want))
		})
	}
}

func validateSpecFromSimpleBundle(t *testing.T, gotSpec *cluster.Spec) {
	VersionsBundle, err := gotSpec.GetCPVersionsBundle()
	if err != nil {
		t.Errorf("Could not get VersionsBundle")
	}
	validateVersionedRepo(t, VersionsBundle.KubeDistro.Kubernetes, "public.ecr.aws/eks-distro/kubernetes", "v1.19.8-eks-1-19-4")
	validateVersionedRepo(t, VersionsBundle.KubeDistro.CoreDNS, "public.ecr.aws/eks-distro/coredns", "v1.8.0-eks-1-19-4")
	validateVersionedRepo(t, VersionsBundle.KubeDistro.Etcd, "public.ecr.aws/eks-distro/etcd-io", "v3.4.14-eks-1-19-4")
	validateImageURI(t, VersionsBundle.KubeDistro.NodeDriverRegistrar, "public.ecr.aws/eks-distro/kubernetes-csi/node-driver-registrar:v2.1.0-eks-1-19-4")
	validateImageURI(t, VersionsBundle.KubeDistro.LivenessProbe, "public.ecr.aws/eks-distro/kubernetes-csi/livenessprobe:v2.2.0-eks-1-19-4")
	validateImageURI(t, VersionsBundle.KubeDistro.ExternalAttacher, "public.ecr.aws/eks-distro/kubernetes-csi/external-attacher:v3.1.0-eks-1-19-4")
	validateImageURI(t, VersionsBundle.KubeDistro.ExternalProvisioner, "public.ecr.aws/eks-distro/kubernetes-csi/external-provisioner:v2.1.1-eks-1-19-4")
	validateImageURI(t, VersionsBundle.KubeDistro.EtcdImage, "public.ecr.aws/eks-distro/etcd-io/etcd:v3.4.14-eks-1-19-4")
	validateImageURI(t, VersionsBundle.KubeDistro.KubeProxy, "public.ecr.aws/eks-distro/kubernetes/kube-proxy:v1.19.8-eks-1-19-4")
	if VersionsBundle.KubeDistro.EtcdVersion != "3.4.14" {
		t.Errorf("GetNewSpec() = Spec: Invalid etcd version, got %s, want 3.4.14", VersionsBundle.KubeDistro.EtcdVersion)
	}
}

func TestKubeDistroImagesErrorEksd(t *testing.T) {
	g := NewWithT(t)
	config := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			Spec: anywherev1.ClusterSpec{
				KubernetesVersion: anywherev1.Kube124,
			},
		},
		OIDCConfigs: map[string]*anywherev1.OIDCConfig{
			"myconfig": {},
		},
		AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{
			"myconfig": {},
		},
	}
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			Number: 2,
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.24",
				},
			},
		},
	}
	eksd := test.EksdReleasesMap()

	spec, _ := cluster.NewSpec(config, bundles, eksd)
	images := spec.KubeDistroImages()
	g.Expect(len(images)).To(BeZero())
}

func TestKubeDistroImagesErrorVersionsBundle(t *testing.T) {
	g := NewWithT(t)
	config := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			Spec: anywherev1.ClusterSpec{
				KubernetesVersion: anywherev1.Kube124,
			},
		},
		OIDCConfigs: map[string]*anywherev1.OIDCConfig{
			"myconfig": {},
		},
		AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{
			"myconfig": {},
		},
	}
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			Number: 2,
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.24",
				},
			},
		},
	}
	eksd := test.EksdReleasesMap()

	spec, _ := cluster.NewSpec(config, bundles, eksd)
	spec.VersionsBundles["1.24"] = nil
	images := spec.KubeDistroImages()
	g.Expect(len(images)).To(BeZero())
}

func TestGetVersionsBundleWorkerError(t *testing.T) {
	g := NewWithT(t)
	kube123 := anywherev1.KubernetesVersion("1.23")
	config := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			Spec: anywherev1.ClusterSpec{
				KubernetesVersion: anywherev1.Kube124,
				WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
					{
						KubernetesVersion: &kube123,
					},
				},
			},
		},
		OIDCConfigs: map[string]*anywherev1.OIDCConfig{
			"myconfig": {},
		},
		AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{
			"myconfig": {},
		},
	}
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			Number: 2,
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.24",
				},
			},
		},
	}
	eksd := test.EksdReleasesMap()

	_, err := cluster.NewSpec(config, bundles, eksd)
	g.Expect(err).ToNot(BeNil())
}

func TestGetVersionsBundleNotFound(t *testing.T) {
	g := NewWithT(t)
	config := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			Spec: anywherev1.ClusterSpec{
				KubernetesVersion: anywherev1.Kube124,
			},
		},
		OIDCConfigs: map[string]*anywherev1.OIDCConfig{
			"myconfig": {},
		},
		AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{
			"myconfig": {},
		},
	}
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			Number: 2,
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.24",
				},
			},
		},
	}
	eksd := test.EksdReleasesMap()

	spec, _ := cluster.NewSpec(config, bundles, eksd)
	_, err := spec.GetVersionBundles("1.22")
	g.Expect(err).ToNot(BeNil())
}

func TestGetVersionsBundlesErrorEksdEmpty(t *testing.T) {
	g := NewWithT(t)
	kube123 := anywherev1.KubernetesVersion("1.23")
	config := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			Spec: anywherev1.ClusterSpec{
				KubernetesVersion: anywherev1.Kube124,
				WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
					{
						KubernetesVersion: &kube123,
					},
				},
			},
		},
		OIDCConfigs: map[string]*anywherev1.OIDCConfig{
			"myconfig": {},
		},
		AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{
			"myconfig": {},
		},
	}
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			Number: 2,
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.24",
				},
			},
		},
	}

	_, err := cluster.NewSpec(config, bundles, map[anywherev1.KubernetesVersion]*eksdv1.Release{})
	g.Expect(err).ToNot(BeNil())
}

func TestGetVersionsBundlesErrorInvalidEksd(t *testing.T) {
	g := NewWithT(t)
	kube123 := anywherev1.KubernetesVersion("1.23")
	config := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			Spec: anywherev1.ClusterSpec{
				KubernetesVersion: anywherev1.Kube124,
				WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
					{
						KubernetesVersion: &kube123,
					},
				},
			},
		},
		OIDCConfigs: map[string]*anywherev1.OIDCConfig{
			"myconfig": {},
		},
		AWSIAMConfigs: map[string]*anywherev1.AWSIamConfig{
			"myconfig": {},
		},
	}
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			Number: 2,
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.24",
				},
			},
		},
	}
	eksd := map[anywherev1.KubernetesVersion]*eksdv1.Release{
		"1.24": {
			TypeMeta: metav1.TypeMeta{
				Kind:       "Release",
				APIVersion: "distro.eks.amazonaws.com/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "eksa-system",
			},
			Spec: eksdv1.ReleaseSpec{
				Number: 1,
			},
			Status: eksdv1.ReleaseStatus{
				Components: []eksdv1.Component{
					{
						Assets: nil,
					},
				},
			},
		},
	}

	_, err := cluster.NewSpec(config, bundles, eksd)
	g.Expect(err).ToNot(BeNil())
}

func TestGetOldAndNewCPVersion(t *testing.T) {
	g := NewWithT(t)
	old := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: &anywherev1.Cluster{
				Spec: anywherev1.ClusterSpec{
					KubernetesVersion: anywherev1.Kube119,
				},
			},
		},
		VersionsBundles: test.VersionsBundlesMap(),
	}
	new := old.DeepCopy()

	_, _, err := cluster.GetOldAndNewCPVersionBundle(old, new)
	g.Expect(err).To(BeNil())
}

func TestGetOldAndNewCPVersionError(t *testing.T) {
	g := NewWithT(t)
	old := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: &anywherev1.Cluster{
				Spec: anywherev1.ClusterSpec{
					KubernetesVersion: anywherev1.Kube119,
				},
			},
		},
		VersionsBundles: make(map[anywherev1.KubernetesVersion]*cluster.VersionsBundle),
	}
	new := old.DeepCopy()

	_, _, err := cluster.GetOldAndNewCPVersionBundle(old, new)
	g.Expect(err).ToNot(BeNil())

	old.VersionsBundles = test.VersionsBundlesMap()
	_, _, err = cluster.GetOldAndNewCPVersionBundle(old, new)
	g.Expect(err).ToNot(BeNil())
}

func validateImageURI(t *testing.T, gotImage releasev1.Image, wantURI string) {
	if gotImage.URI != wantURI {
		t.Errorf("GetNewSpec() = Spec: Invalid kubernetes URI, got %s, want %s", gotImage.URI, wantURI)
	}
}

func validateVersionedRepo(t *testing.T, gotImage cluster.VersionedRepository, wantRepo, wantTag string) {
	if gotImage.Repository != wantRepo {
		t.Errorf("GetNewSpec() = Spec: Invalid kubernetes repo, got %s, want %s", gotImage.Repository, wantRepo)
	}
	if gotImage.Tag != wantTag {
		t.Errorf("GetNewSpec() = Spec: Invalid kubernetes repo, got %s, want %s", gotImage.Tag, wantTag)
	}
}

func readEksdRelease(tb testing.TB, url string) *eksdv1.Release {
	tb.Helper()
	r := files.NewReader()
	release, err := eksd.ReadManifest(r, url)
	if err != nil {
		tb.Fatalf("Failed reading eks-d manifest: %s", err)
	}

	return release
}
