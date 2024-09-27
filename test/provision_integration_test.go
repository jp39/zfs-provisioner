//go:build integration

package test

import (
	"context"
	"flag"
	"k8s.io/klog/v2"
	"math/rand"
	"os"
	"strconv"
	"testing"

	gozfs "github.com/mistifyio/go-zfs/v3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v10/controller"

	"github.com/jp39/zfs-provisioner/pkg/provisioner"
	"github.com/jp39/zfs-provisioner/pkg/zfs"
)

var (
	parentDataset = flag.String("parentDataset", "", "parent dataset")
)

type ProvisionTestSuit struct {
	suite.Suite
	p               *provisioner.ZFSProvisioner
	datasetPrefix   string
	createdDatasets []string
}

func TestProvisionSuite(t *testing.T) {
	s := ProvisionTestSuit{
		datasetPrefix:   "pv-test-" + strconv.Itoa(rand.Int()),
		createdDatasets: make([]string, 0),
	}
	suite.Run(t, &s)
}

func (suite *ProvisionTestSuit) SetupSuite() {
	path := os.Getenv("PATH")
	pwd, _ := os.Getwd()
	err := os.Setenv("PATH", pwd+":"+path)
	log := klog.NewKlogr()
	require.NoError(suite.T(), err)
	prov, err := provisioner.NewZFSProvisioner("pv.kubernetes.io/zfs", *parentDataset, log)
	require.NoError(suite.T(), err)
	suite.p = prov
}

func (suite *ProvisionTestSuit) TearDownSuite() {
	for _, dataset := range suite.createdDatasets {
		err := zfs.NewInterface().DestroyDataset(&zfs.Dataset{
			Name:     *parentDataset + "/" + dataset,
		}, zfs.DestroyRecursively)
		require.NoError(suite.T(), err)
	}
}

func (suite *ProvisionTestSuit) TestDefaultProvisionDataset() {
	dataset := provisionDataset(suite, "default", map[string]string{})
	assertZfsReservation(suite.T(), dataset, true)
}

func (suite *ProvisionTestSuit) TestThickProvisionDataset() {
	dataset := provisionDataset(suite, "thick", map[string]string{
		provisioner.ReserveSpaceParameter:    "true",
	})
	assertZfsReservation(suite.T(), dataset, true)
}

func (suite *ProvisionTestSuit) TestThinProvisionDataset() {
	dataset := provisionDataset(suite, "thin", map[string]string{
		provisioner.ReserveSpaceParameter:    "false",
	})
	assertZfsReservation(suite.T(), dataset, false)
}

func provisionDataset(suite *ProvisionTestSuit, name string, parameters map[string]string) string {
	t := suite.T()
	pvName := suite.datasetPrefix + "_" + name
	fullDataset := *parentDataset + "/" + pvName
	datasetDirectory := "/" + fullDataset
	policy := v1.PersistentVolumeReclaimRetain
	options := controller.ProvisionOptions{
		PVName: pvName,
		PVC:    newClaim(resource.MustParse("10M"), []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce, v1.ReadOnlyMany}),
		StorageClass: &storagev1.StorageClass{
			Parameters:    parameters,
			ReclaimPolicy: &policy,
		},
	}

	_, _, err := suite.p.Provision(context.Background(), options)
	suite.createdDatasets = append(suite.createdDatasets, pvName)
	assert.NoError(t, err)
	require.DirExists(t, datasetDirectory)
	return fullDataset
}

func assertZfsReservation(t *testing.T, datasetName string, reserve bool) {
	dataset, err := gozfs.GetDataset(datasetName)
	assert.NoError(t, err)

	refreserved, err := dataset.GetProperty("refreservation")
	assert.NoError(t, err)

	refquota, err := dataset.GetProperty("refquota")
	assert.NoError(t, err)

	if reserve {
		assert.Equal(t, refquota, refreserved)
	} else {
		assert.Equal(t, "none", refreserved)
	}
}

func newClaim(capacity resource.Quantity, accessModes []v1.PersistentVolumeAccessMode) *v1.PersistentVolumeClaim {
	storageClassName := "zfs"
	claim := &v1.PersistentVolumeClaim{
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: accessModes,
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: capacity,
				},
			},
			StorageClassName: &storageClassName,
		},
	}
	return claim
}
