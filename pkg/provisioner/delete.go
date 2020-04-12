package provisioner

import (
	"fmt"
	"github.com/ccremer/kubernetes-zfs-provisioner/pkg/zfs"
	core "k8s.io/api/core/v1"
	"k8s.io/klog"
)

// Delete removes a given volume from the server
func (p *ZFSProvisioner) Delete(volume *core.PersistentVolume) error {
	for _, annotation := range []string{DatasetPathAnnotation, ZFSHostAnnotation} {
		value := volume.ObjectMeta.Annotations[annotation]
		if value == "" {
			return fmt.Errorf("annotation '%s' not found or empty, cannot determine which ZFS dataset to destroy", annotation)
		}
	}
	datasetPath := volume.ObjectMeta.Annotations[DatasetPathAnnotation]
	zfsHost := volume.ObjectMeta.Annotations[ZFSHostAnnotation]

	klog.V(3).Info("acquiring lock...")
	globalLock.Lock()
	defer globalLock.Unlock()
	if err := setEnvironmentVars(zfsHost, false, datasetPath); err != nil {
		return err
	}

	err := p.zfs.DestroyDataset(&zfs.Dataset{Name: datasetPath}, zfs.DestroyRecursively)
	if err != nil {
		return fmt.Errorf("error destroying dataset: %w", err)
	}

	klog.Infof("dataset \"%s\": destroyed", datasetPath)
	return nil
}
