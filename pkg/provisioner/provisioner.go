package provisioner

import (
	"fmt"
	"strings"

	"github.com/jp39/zfs-provisioner/pkg/zfs"
	"k8s.io/klog/v2"
)

const (
	DatasetPathAnnotation = "zfs.pv.kubernetes.io/zfs-dataset-path"

	RefQuotaProperty       = "refquota"
	RefReservationProperty = "refreservation"
	ShareNfsProperty       = "sharenfs"
	ManagedByProperty      = "io.kubernetes.pv.zfs:managed_by"
	ReclaimPolicyProperty  = "io.kubernetes.pv.zfs:reclaim_policy"
)

// ZFSProvisioner implements the Provisioner interface to create and export ZFS volumes
type ZFSProvisioner struct {
	zfs           zfs.Interface
	log           klog.Logger
	InstanceName  string
	ParentDataset string
}

// NewZFSProvisioner returns a new ZFSProvisioner based on a given optional
// zap.Logger. If none given, zaps default production logger is used.
func NewZFSProvisioner(instanceName string, parentDataset string, logger klog.Logger) (*ZFSProvisioner, error) {
	if strings.HasPrefix(parentDataset, "/") || strings.HasSuffix(parentDataset, "/") {
		return nil, fmt.Errorf("parentDataset must not begin or end with '/': %s", parentDataset)
	}

	return &ZFSProvisioner{
		log:           logger,
		zfs:           zfs.NewInterface(),
		InstanceName:  instanceName,
		ParentDataset: parentDataset,
	}, nil
}
