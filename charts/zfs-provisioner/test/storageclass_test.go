package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1"
)

var tplStorageclass = []string{"templates/storageclass.yaml"}

func Test_Storageclass_GivenClassesEnabled_WhenNoPolicyDefined_ThenRenderDefault(t *testing.T) {
	options := &helm.Options{
		ValuesFiles: []string{"values/storageclass_1.yaml"},
		SetValues: map[string]string{
			"storageClass.classes[0].policy": "",
		},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, tplStorageclass)

	var class v1.StorageClass
	helm.UnmarshalK8SYaml(t, output, &class)

	expectedPolicy := core.PersistentVolumeReclaimDelete
	assert.Equal(t, &expectedPolicy, class.ReclaimPolicy)
}

func Test_StorageClass_GivenClassesEnabled_WhenReserveSpaceUndefined_ThenRenderDefault(t *testing.T) {
	options := &helm.Options{
		ValuesFiles: []string{"values/storageclass_1.yaml"},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, tplStorageclass)

	var class v1.StorageClass
	helm.UnmarshalK8SYaml(t, output, &class)

	value, exists := class.Parameters["reserveSpace"]
	assert.False(t, exists)
	assert.Empty(t, value)
}

func Test_StorageClass_GivenClassesEnabled_WhenReserveSpaceFalse_ThenRenderReserveSpace(t *testing.T) {
	options := &helm.Options{
		SetValues: map[string]string{
			"storageClass.create":                  "true",
			"storageClass.classes[0].reserveSpace": "false",
		},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, tplStorageclass)

	var class v1.StorageClass
	helm.UnmarshalK8SYaml(t, output, &class)

	assert.Equal(t, "false", class.Parameters["reserveSpace"])
}

func Test_StorageClass_GivenClassesEnabled_WhenReserveSpaceTrue_ThenRenderReserveSpace(t *testing.T) {
	options := &helm.Options{
		SetValues: map[string]string{
			"storageClass.create":                  "true",
			"storageClass.classes[0].reserveSpace": "true",
		},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, tplStorageclass)

	var class v1.StorageClass
	helm.UnmarshalK8SYaml(t, output, &class)

	assert.Equal(t, "true", class.Parameters["reserveSpace"])
}
