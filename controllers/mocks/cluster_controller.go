// Code generated by MockGen. DO NOT EDIT.
// Source: controllers/cluster_controller.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	v1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	controller "github.com/aws/eks-anywhere/pkg/controller"
	clusters "github.com/aws/eks-anywhere/pkg/controller/clusters"
	curatedpackages "github.com/aws/eks-anywhere/pkg/curatedpackages"
	registrymirror "github.com/aws/eks-anywhere/pkg/registrymirror"
	v1alpha10 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	logr "github.com/go-logr/logr"
	gomock "github.com/golang/mock/gomock"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockPackagesClient is a mock of PackagesClient interface.
type MockPackagesClient struct {
	ctrl     *gomock.Controller
	recorder *MockPackagesClientMockRecorder
}

// MockPackagesClientMockRecorder is the mock recorder for MockPackagesClient.
type MockPackagesClientMockRecorder struct {
	mock *MockPackagesClient
}

// NewMockPackagesClient creates a new mock instance.
func NewMockPackagesClient(ctrl *gomock.Controller) *MockPackagesClient {
	mock := &MockPackagesClient{ctrl: ctrl}
	mock.recorder = &MockPackagesClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPackagesClient) EXPECT() *MockPackagesClientMockRecorder {
	return m.recorder
}

// EnableFullLifecycle mocks base method.
func (m *MockPackagesClient) EnableFullLifecycle(ctx context.Context, log logr.Logger, clusterName, kubeConfig string, chart *v1alpha10.Image, registry *registrymirror.RegistryMirror, options ...curatedpackages.PackageControllerClientOpt) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, log, clusterName, kubeConfig, chart, registry}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "EnableFullLifecycle", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// EnableFullLifecycle indicates an expected call of EnableFullLifecycle.
func (mr *MockPackagesClientMockRecorder) EnableFullLifecycle(ctx, log, clusterName, kubeConfig, chart, registry interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, log, clusterName, kubeConfig, chart, registry}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnableFullLifecycle", reflect.TypeOf((*MockPackagesClient)(nil).EnableFullLifecycle), varargs...)
}

// Reconcile mocks base method.
func (m *MockPackagesClient) Reconcile(arg0 context.Context, arg1 logr.Logger, arg2 client.Client, arg3 *v1alpha1.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reconcile", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reconcile indicates an expected call of Reconcile.
func (mr *MockPackagesClientMockRecorder) Reconcile(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconcile", reflect.TypeOf((*MockPackagesClient)(nil).Reconcile), arg0, arg1, arg2, arg3)
}

// ReconcileDelete mocks base method.
func (m *MockPackagesClient) ReconcileDelete(arg0 context.Context, arg1 logr.Logger, arg2 curatedpackages.KubeDeleter, arg3 *v1alpha1.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReconcileDelete", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReconcileDelete indicates an expected call of ReconcileDelete.
func (mr *MockPackagesClientMockRecorder) ReconcileDelete(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReconcileDelete", reflect.TypeOf((*MockPackagesClient)(nil).ReconcileDelete), arg0, arg1, arg2, arg3)
}

// MockProviderClusterReconcilerRegistry is a mock of ProviderClusterReconcilerRegistry interface.
type MockProviderClusterReconcilerRegistry struct {
	ctrl     *gomock.Controller
	recorder *MockProviderClusterReconcilerRegistryMockRecorder
}

// MockProviderClusterReconcilerRegistryMockRecorder is the mock recorder for MockProviderClusterReconcilerRegistry.
type MockProviderClusterReconcilerRegistryMockRecorder struct {
	mock *MockProviderClusterReconcilerRegistry
}

// NewMockProviderClusterReconcilerRegistry creates a new mock instance.
func NewMockProviderClusterReconcilerRegistry(ctrl *gomock.Controller) *MockProviderClusterReconcilerRegistry {
	mock := &MockProviderClusterReconcilerRegistry{ctrl: ctrl}
	mock.recorder = &MockProviderClusterReconcilerRegistryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProviderClusterReconcilerRegistry) EXPECT() *MockProviderClusterReconcilerRegistryMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockProviderClusterReconcilerRegistry) Get(datacenterKind string) clusters.ProviderClusterReconciler {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", datacenterKind)
	ret0, _ := ret[0].(clusters.ProviderClusterReconciler)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockProviderClusterReconcilerRegistryMockRecorder) Get(datacenterKind interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockProviderClusterReconcilerRegistry)(nil).Get), datacenterKind)
}

// MockAWSIamConfigReconciler is a mock of AWSIamConfigReconciler interface.
type MockAWSIamConfigReconciler struct {
	ctrl     *gomock.Controller
	recorder *MockAWSIamConfigReconcilerMockRecorder
}

// MockAWSIamConfigReconcilerMockRecorder is the mock recorder for MockAWSIamConfigReconciler.
type MockAWSIamConfigReconcilerMockRecorder struct {
	mock *MockAWSIamConfigReconciler
}

// NewMockAWSIamConfigReconciler creates a new mock instance.
func NewMockAWSIamConfigReconciler(ctrl *gomock.Controller) *MockAWSIamConfigReconciler {
	mock := &MockAWSIamConfigReconciler{ctrl: ctrl}
	mock.recorder = &MockAWSIamConfigReconcilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAWSIamConfigReconciler) EXPECT() *MockAWSIamConfigReconcilerMockRecorder {
	return m.recorder
}

// EnsureCASecret mocks base method.
func (m *MockAWSIamConfigReconciler) EnsureCASecret(ctx context.Context, logger logr.Logger, cluster *v1alpha1.Cluster) (controller.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureCASecret", ctx, logger, cluster)
	ret0, _ := ret[0].(controller.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureCASecret indicates an expected call of EnsureCASecret.
func (mr *MockAWSIamConfigReconcilerMockRecorder) EnsureCASecret(ctx, logger, cluster interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureCASecret", reflect.TypeOf((*MockAWSIamConfigReconciler)(nil).EnsureCASecret), ctx, logger, cluster)
}

// Reconcile mocks base method.
func (m *MockAWSIamConfigReconciler) Reconcile(ctx context.Context, logger logr.Logger, cluster *v1alpha1.Cluster) (controller.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reconcile", ctx, logger, cluster)
	ret0, _ := ret[0].(controller.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Reconcile indicates an expected call of Reconcile.
func (mr *MockAWSIamConfigReconcilerMockRecorder) Reconcile(ctx, logger, cluster interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reconcile", reflect.TypeOf((*MockAWSIamConfigReconciler)(nil).Reconcile), ctx, logger, cluster)
}

// ReconcileDelete mocks base method.
func (m *MockAWSIamConfigReconciler) ReconcileDelete(ctx context.Context, logger logr.Logger, cluster *v1alpha1.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReconcileDelete", ctx, logger, cluster)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReconcileDelete indicates an expected call of ReconcileDelete.
func (mr *MockAWSIamConfigReconcilerMockRecorder) ReconcileDelete(ctx, logger, cluster interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReconcileDelete", reflect.TypeOf((*MockAWSIamConfigReconciler)(nil).ReconcileDelete), ctx, logger, cluster)
}

// MockClusterValidator is a mock of ClusterValidator interface.
type MockClusterValidator struct {
	ctrl     *gomock.Controller
	recorder *MockClusterValidatorMockRecorder
}

// MockClusterValidatorMockRecorder is the mock recorder for MockClusterValidator.
type MockClusterValidatorMockRecorder struct {
	mock *MockClusterValidator
}

// NewMockClusterValidator creates a new mock instance.
func NewMockClusterValidator(ctrl *gomock.Controller) *MockClusterValidator {
	mock := &MockClusterValidator{ctrl: ctrl}
	mock.recorder = &MockClusterValidatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClusterValidator) EXPECT() *MockClusterValidatorMockRecorder {
	return m.recorder
}

// ValidateEksaVersionExists mocks base method.
func (m *MockClusterValidator) ValidateEksaVersionExists(ctx context.Context, log logr.Logger, cluster *v1alpha1.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateEksaVersionExists", ctx, log, cluster)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateEksaVersionExists indicates an expected call of ValidateEksaVersionExists.
func (mr *MockClusterValidatorMockRecorder) ValidateEksaVersionExists(ctx, log, cluster interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateEksaVersionExists", reflect.TypeOf((*MockClusterValidator)(nil).ValidateEksaVersionExists), ctx, log, cluster)
}

// ValidateManagementClusterEksaVersion mocks base method.
func (m *MockClusterValidator) ValidateManagementClusterEksaVersion(ctx context.Context, log logr.Logger, cluster *v1alpha1.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateManagementClusterEksaVersion", ctx, log, cluster)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateManagementClusterEksaVersion indicates an expected call of ValidateManagementClusterEksaVersion.
func (mr *MockClusterValidatorMockRecorder) ValidateManagementClusterEksaVersion(ctx, log, cluster interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateManagementClusterEksaVersion", reflect.TypeOf((*MockClusterValidator)(nil).ValidateManagementClusterEksaVersion), ctx, log, cluster)
}

// ValidateManagementClusterName mocks base method.
func (m *MockClusterValidator) ValidateManagementClusterName(ctx context.Context, log logr.Logger, cluster *v1alpha1.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateManagementClusterName", ctx, log, cluster)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateManagementClusterName indicates an expected call of ValidateManagementClusterName.
func (mr *MockClusterValidatorMockRecorder) ValidateManagementClusterName(ctx, log, cluster interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateManagementClusterName", reflect.TypeOf((*MockClusterValidator)(nil).ValidateManagementClusterName), ctx, log, cluster)
}

// ValidateManagementWorkloadSkew mocks base method.
func (m *MockClusterValidator) ValidateManagementWorkloadSkew(ctx context.Context, log logr.Logger, mgmtCluster *v1alpha1.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateManagementWorkloadSkew", ctx, log, mgmtCluster)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateManagementWorkloadSkew indicates an expected call of ValidateManagementWorkloadSkew.
func (mr *MockClusterValidatorMockRecorder) ValidateManagementWorkloadSkew(ctx, log, mgmtCluster interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateManagementWorkloadSkew", reflect.TypeOf((*MockClusterValidator)(nil).ValidateManagementWorkloadSkew), ctx, log, mgmtCluster)
}
