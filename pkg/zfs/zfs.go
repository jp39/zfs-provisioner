package zfs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	gozfs "github.com/mistifyio/go-zfs/v3"
	"k8s.io/klog/v2"
)

type (
	// Interface abstracts the underlying ZFS library with the minimum functionality implemented
	Interface interface {
		GetDataset(name string) (*Dataset, error)
		CreateDataset(name string, properties map[string]string) (*Dataset, error)
		DestroyDataset(dataset *Dataset, flag DestroyFlag) error
		SetPermissions(dataset *Dataset) error
	}
	// Dataset is a representation of a ZFS dataset
	Dataset struct {
		datasetImpl *gozfs.Dataset

		Name       string
		Mountpoint string
	}
	DestroyFlag int
	zfsImpl     struct{}
)

const (
	DestroyRecursively DestroyFlag = 2
	HostEnvVar                     = "ZFS_HOST"
)

var (
	globalLock = sync.Mutex{}
)

func (z *zfsImpl) GetDataset(name string) (*Dataset, error) {
	dataset, err := gozfs.GetDataset(name)
	if err != nil {
		return nil, err
	}
	return &Dataset{
		datasetImpl: dataset,
		Name:        dataset.Name,
		Mountpoint:  dataset.Mountpoint,
	}, err
}

func (z *zfsImpl) CreateDataset(name string, properties map[string]string) (*Dataset, error) {
	klog.V(3).InfoS("creating dataset", "name", name)
	dataset, err := gozfs.CreateFilesystem(name, properties)
	if err != nil {
		return nil, err
	}
	return &Dataset{
		datasetImpl: dataset,
		Name:        dataset.Name,
		Mountpoint:  dataset.Mountpoint,
	}, err
}

func (z *zfsImpl) DestroyDataset(dataset *Dataset, flag DestroyFlag) error {
	if err := validateDataset(dataset); err != nil {
		return err
	}
	if dataset.datasetImpl == nil {
		ds, err := z.GetDataset(dataset.Name)
		if err != nil {
			return err
		}
		dataset.datasetImpl = ds.datasetImpl
	}
	var destrFlag gozfs.DestroyFlag
	switch flag {
	case DestroyRecursively:
		destrFlag = gozfs.DestroyRecursive
		break
	default:
		return fmt.Errorf("programmer error: flag not implemented: %d", flag)
	}
	return dataset.datasetImpl.Destroy(destrFlag)
}

func (z *zfsImpl) SetPermissions(dataset *Dataset) error {
	if err := validateDataset(dataset); err != nil {
		return err
	}
	if dataset.Mountpoint == "" {
		return fmt.Errorf("undefined mountpoint for dataset: %s", dataset.Name)
	}

	cmd := exec.Command("update-permissions", dataset.Mountpoint)
	if filepath.IsAbs(cmd.Path) {
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("could not update permissions on '%s': %w: %s", dataset.Mountpoint, err, out)
		}
		return nil
	}

	// update-permissions executable not found in PATH
	st, err := os.Lstat(dataset.Mountpoint)
	if err != nil {
		return err
	}

	// Add group write bit
	if err := os.Chmod(dataset.Mountpoint, st.Mode()|0o020); err != nil {
		return err
	}

	return nil
}

func NewInterface() *zfsImpl {
	return &zfsImpl{}
}

func setEnvironmentVars(hostName string) error {
	return os.Setenv(HostEnvVar, hostName)
}

func validateDataset(dataset *Dataset) error {
	if dataset.Name == "" {
		return errors.New("undefined dataset name")
	}
	return nil
}
