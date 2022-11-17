package reconciler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	clusterspec "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/providers/docker/reconciler"
	dockereconcilermocks "github.com/aws/eks-anywhere/pkg/providers/docker/reconciler/mocks"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

const (
	clusterNamespace = "test-namespace"
)

func TestReconcilerReconcileSuccess(t *testing.T) {
	tt := newReconcilerTest(t)

	logger := test.NewNullLogger()

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcilerReconcileWorkerNodesSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.cluster.Name = "my-management-cluster"
	tt.cluster.SetSelfManaged()
	capiCluster := capiCluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})

	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)
	tt.createAllObjs()

	logger := test.NewNullLogger()

	result, err := tt.reconciler().ReconcileWorkerNodes(tt.ctx, logger, tt.cluster)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	tt.ShouldEventuallyExist(tt.ctx,
		&clusterv1.MachineDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-management-cluster-md-0",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)

	tt.ShouldEventuallyExist(tt.ctx,
		&bootstrapv1.KubeadmConfigTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-management-cluster-md-0-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
}

func TestReconcileCNISuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.withFakeClient()

	logger := test.NewNullLogger()
	remoteClient := fake.NewClientBuilder().Build()
	spec := tt.buildSpec()

	tt.remoteClientRegistry.EXPECT().GetClient(
		tt.ctx, client.ObjectKey{Name: "workload-cluster", Namespace: "eksa-system"},
	).Return(remoteClient, nil)
	tt.cniReconciler.EXPECT().Reconcile(tt.ctx, logger, remoteClient, spec)

	result, err := tt.reconciler().ReconcileCNI(tt.ctx, logger, spec)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcileCNIErrorClientRegistry(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.withFakeClient()

	logger := test.NewNullLogger()
	spec := tt.buildSpec()

	tt.remoteClientRegistry.EXPECT().GetClient(
		tt.ctx, client.ObjectKey{Name: "workload-cluster", Namespace: "eksa-system"},
	).Return(nil, errors.New("building client"))

	result, err := tt.reconciler().ReconcileCNI(tt.ctx, logger, spec)

	tt.Expect(err).To(MatchError(ContainSubstring("building client")))
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))
}

type reconcilerTest struct {
	t testing.TB
	*WithT
	*envtest.APIExpecter
	ctx                  context.Context
	cniReconciler        *dockereconcilermocks.MockCNIReconciler
	remoteClientRegistry *dockereconcilermocks.MockRemoteClientRegistry
	cluster              *anywherev1.Cluster
	client               client.Client
	env                  *envtest.Environment
	eksaSupportObjs      []client.Object
	datacenterConfig     *anywherev1.DockerDatacenterConfig
}

func newReconcilerTest(t testing.TB) *reconcilerTest {
	ctrl := gomock.NewController(t)
	cniReconciler := dockereconcilermocks.NewMockCNIReconciler(ctrl)
	remoteClientRegistry := dockereconcilermocks.NewMockRemoteClientRegistry(ctrl)
	c := env.Client()

	bundle := test.Bundle()

	managementCluster := dockerCluster(func(c *anywherev1.Cluster) {
		c.Name = "management-cluster"
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: c.Name,
		}
		c.Spec.BundlesRef = &anywherev1.BundlesRef{
			Name:       bundle.Name,
			Namespace:  bundle.Namespace,
			APIVersion: bundle.APIVersion,
		}
	})

	workloadClusterDatacenter := dataCenter()

	cluster := dockerCluster(func(c *anywherev1.Cluster) {
		c.Name = "workload-cluster"
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: managementCluster.Name,
		}
		c.Spec.BundlesRef = &anywherev1.BundlesRef{
			Name:       bundle.Name,
			Namespace:  bundle.Namespace,
			APIVersion: bundle.APIVersion,
		}
		c.Spec.ControlPlaneConfiguration = anywherev1.ControlPlaneConfiguration{
			Count: 1,
		}
		c.Spec.DatacenterRef = anywherev1.Ref{
			Kind: anywherev1.DockerDatacenterKind,
			Name: workloadClusterDatacenter.Name,
		}

		c.Spec.WorkerNodeGroupConfigurations = append(c.Spec.WorkerNodeGroupConfigurations,
			anywherev1.WorkerNodeGroupConfiguration{
				Count:  ptr.Int(1),
				Name:   "md-0",
				Labels: nil,
			},
		)
	})

	tt := &reconcilerTest{
		t:                    t,
		WithT:                NewWithT(t),
		APIExpecter:          envtest.NewAPIExpecter(t, c),
		ctx:                  context.Background(),
		cniReconciler:        cniReconciler,
		remoteClientRegistry: remoteClientRegistry,
		cluster:              cluster,
		client:               c,
		env:                  env,
		eksaSupportObjs: []client.Object{
			test.Namespace(clusterNamespace),
			test.Namespace(constants.EksaSystemNamespace),
			managementCluster,
			workloadClusterDatacenter,
			bundle,
			test.EksdRelease(),
		},
		datacenterConfig: workloadClusterDatacenter,
	}

	t.Cleanup(tt.cleanup)
	return tt
}

func (tt *reconcilerTest) cleanup() {
	tt.DeleteAndWait(tt.ctx, tt.allObjs()...)
}

func (tt *reconcilerTest) reconciler() *reconciler.Reconciler {
	return reconciler.New(tt.client, tt.cniReconciler, tt.remoteClientRegistry)
}

func (tt *reconcilerTest) buildSpec() *clusterspec.Spec {
	tt.t.Helper()
	spec, err := clusterspec.BuildSpec(tt.ctx, clientutil.NewKubeClient(tt.client), tt.cluster)
	tt.Expect(err).NotTo(HaveOccurred())

	return spec
}

func (tt *reconcilerTest) withFakeClient() {
	tt.client = fake.NewClientBuilder().WithObjects(clientutil.ObjectsToClientObjects(tt.allObjs())...).Build()
}

func (tt *reconcilerTest) createAllObjs() {
	tt.t.Helper()
	envtest.CreateObjs(tt.ctx, tt.t, tt.client, tt.allObjs()...)
}

func (tt *reconcilerTest) allObjs() []client.Object {
	objs := make([]client.Object, 0, len(tt.eksaSupportObjs)+1)
	objs = append(objs, tt.eksaSupportObjs...)
	objs = append(objs, tt.cluster)

	return objs
}

type clusterOpt func(*anywherev1.Cluster)

func dockerCluster(opts ...clusterOpt) *anywherev1.Cluster {
	c := &anywherev1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterNamespace,
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "1.20",
			ClusterNetwork: anywherev1.ClusterNetwork{
				Pods: anywherev1.Pods{
					CidrBlocks: []string{"0.0.0.0"},
				},
				Services: anywherev1.Services{
					CidrBlocks: []string{"0.0.0.0"},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func dataCenter() *anywherev1.DockerDatacenterConfig {
	return &anywherev1.DockerDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.DockerDatacenterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "datacenter",
			Namespace: clusterNamespace,
		},
	}
}

type capiClusterOpt func(*clusterv1.Cluster)

func capiCluster(opts ...capiClusterOpt) *clusterv1.Cluster {
	c := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: clusterv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
		Status: clusterv1.ClusterStatus{
			Conditions: clusterv1.Conditions{
				{
					Type:               clusterapi.ControlPlaneReadyCondition,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(time.Now()),
				},
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
