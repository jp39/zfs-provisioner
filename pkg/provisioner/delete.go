package provisioner

import (
	"context"
	"fmt"

	"github.com/jp39/zfs-provisioner/pkg/zfs"
	core "k8s.io/api/core/v1"
)

// Delete removes a given volume from the server
func (p *ZFSProvisioner) Delete(ctx context.Context, volume *core.PersistentVolume) error {
	for _, annotation := range []string{DatasetPathAnnotation} {
		value := volume.ObjectMeta.Annotations[annotation]
		if value == "" {
			return fmt.Errorf("annotation '%s' not found or empty, cannot determine which ZFS dataset to destroy", annotation)
		}
	}
	datasetPath := volume.ObjectMeta.Annotations[DatasetPathAnnotation]

	err := p.zfs.DestroyDataset(&zfs.Dataset{Name: datasetPath}, zfs.DestroyRecursively)
	if err != nil {
		return fmt.Errorf("error destroying dataset: %w", err)
	}

	p.log.Info("dataset destroyed", "dataset", datasetPath)
	return nil
}
