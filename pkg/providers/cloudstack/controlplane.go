package cloudstack

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta2"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	yamlcapi "github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

// BaseControlPlane represents a CAPI CloudStack control plane.
type BaseControlPlane = clusterapi.ControlPlane[*cloudstackv1.CloudStackCluster, *cloudstackv1.CloudStackMachineTemplate]

// ControlPlane holds the CloudStack specific objects for a CAPI CloudStack control plane.
type ControlPlane struct {
	BaseControlPlane
	Secrets *corev1.Secret
}

// Objects returns the control plane objects associated with the CloudStack cluster.
func (p ControlPlane) Objects() []kubernetes.Object {
	o := p.BaseControlPlane.Objects()
	o = append(o, p.Secrets)

	return o
}

// ControlPlaneSpec builds a CloudStack ControlPlane definition based on an eks-a cluster spec.
func ControlPlaneSpec(ctx context.Context, logger logr.Logger, client kubernetes.Client, clusterSpec *cluster.Spec) (*ControlPlane, error) {
	templateBuilder, err := generateTemplateBuilder(clusterSpec)
	if err != nil {
		return nil, errors.Wrap(err, "generating cloudstack template builder")
	}

	controlPlaneYaml, err := templateBuilder.GenerateCAPISpecControlPlane(
		clusterSpec,
		func(values map[string]interface{}) {
			values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(clusterSpec.Cluster)
			values["etcdTemplateName"] = clusterapi.EtcdMachineTemplateName(clusterSpec.Cluster)
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "generating cloudstack control plane yaml spec")
	}

	parser, builder, err := newControlPlaneParser(logger)
	if err != nil {
		return nil, err
	}

	err = parser.Parse(controlPlaneYaml, builder)
	if err != nil {
		return nil, errors.Wrap(err, "parsing cloudstack control plane yaml")
	}

	cp := builder.ControlPlane
	if err = cp.UpdateImmutableObjectNames(ctx, client, GetMachineTemplate, machineTemplateEqual); err != nil {
		return nil, errors.Wrap(err, "updating cloudstack immutable object names")
	}

	return cp, nil
}

// ControlPlaneBuilder defines the builder for all objects in the CAPI CloudStack control plane.
type controlPlaneBuilder struct {
	BaseBuilder  *yamlcapi.ControlPlaneBuilder[*cloudstackv1.CloudStackCluster, *cloudstackv1.CloudStackMachineTemplate]
	ControlPlane *ControlPlane
}

// BuildFromParsed implements the base yamlcapi.BuildFromParsed and processes any additional objects (secrets) for the CloudStack control plane.
func (b *controlPlaneBuilder) BuildFromParsed(lookup yamlutil.ObjectLookup) error {
	if err := b.BaseBuilder.BuildFromParsed(lookup); err != nil {
		return err
	}

	b.ControlPlane.BaseControlPlane = *b.BaseBuilder.ControlPlane
	for _, obj := range lookup {
		if obj.GetObjectKind().GroupVersionKind().Kind == constants.SecretKind {
			b.ControlPlane.Secrets = obj.(*corev1.Secret)
		}
	}

	return nil
}

func newControlPlaneParser(logger logr.Logger) (*yamlutil.Parser, *controlPlaneBuilder, error) {
	parser, baseBuilder, err := yamlcapi.NewControlPlaneParserAndBuilder(
		logger,
		yamlutil.NewMapping(
			"CloudStackCluster",
			func() *cloudstackv1.CloudStackCluster {
				return &cloudstackv1.CloudStackCluster{}
			},
		),
		yamlutil.NewMapping(
			"CloudStackMachineTemplate",
			func() *cloudstackv1.CloudStackMachineTemplate {
				return &cloudstackv1.CloudStackMachineTemplate{}
			},
		),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "building cloudstack control plane parser")
	}

	err = parser.RegisterMappings(
		yamlutil.NewMapping(constants.SecretKind, func() yamlutil.APIObject {
			return &corev1.Secret{}
		}),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "registering cloudstack control plane mappings in parser")
	}

	builder := &controlPlaneBuilder{
		BaseBuilder:  baseBuilder,
		ControlPlane: &ControlPlane{},
	}

	return parser, builder, nil
}
