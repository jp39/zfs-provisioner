package provisioner

import (
	"github.com/stretchr/testify/mock"

	"github.com/jp39/kubernetes-zfs-provisioner/pkg/zfs"
)

type (
	zfsStub struct {
		mock.Mock
	}
)

func (z *zfsStub) GetDataset(name string) (*zfs.Dataset, error) {
	args := z.Called(name)
	return args.Get(0).(*zfs.Dataset), args.Error(1)
}

func (z *zfsStub) CreateDataset(name string, properties map[string]string) (*zfs.Dataset, error) {
	args := z.Called(name, properties)
	return args.Get(0).(*zfs.Dataset), args.Error(1)
}

func (z *zfsStub) DestroyDataset(dataset *zfs.Dataset, flag zfs.DestroyFlag) error {
	args := z.Called(dataset, flag)
	return args.Error(0)
}

func (z *zfsStub) SetPermissions(dataset *zfs.Dataset) error {
	args := z.Called(dataset)
	return args.Error(0)
}

func NewZFSProvisionerStub(stub *zfsStub) (*ZFSProvisioner, error) {
	return &ZFSProvisioner{
		zfs:          stub,
		InstanceName: "test",
	}, nil
}
