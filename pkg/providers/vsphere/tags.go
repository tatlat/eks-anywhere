package vsphere

import (
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func requiredTemplateTags(clusterSpec *cluster.Spec, machineConfig *v1alpha1.VSphereMachineConfig) []string {
	tagsByCategory := requiredTemplateTagsByCategory(clusterSpec, machineConfig)
	tags := make([]string, 0, len(tagsByCategory))
	for _, t := range tagsByCategory {
		tags = append(tags, t...)
	}

	return tags
}

func requiredTemplateTagsByCategory(clusterSpec *cluster.Spec, machineConfig *v1alpha1.VSphereMachineConfig) map[string][]string {
	osFamily := machineConfig.Spec.OSFamily
	vb, err := clusterSpec.GetCPVersionsBundle()
	if err != nil {
		return nil
	}
	return map[string][]string{
		"eksdRelease": {fmt.Sprintf("eksdRelease:%s", vb.EksD.Name)},
		"os":          {fmt.Sprintf("os:%s", strings.ToLower(string(osFamily)))},
	}
}
