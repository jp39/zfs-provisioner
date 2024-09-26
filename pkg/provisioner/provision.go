package provisioner

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strconv"

	"github.com/jp39/zfs-provisioner/pkg/zfs"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v10/controller"
)

// Provision creates a PersistentVolume, sets quota.
func (p *ZFSProvisioner) Provision(ctx context.Context, options controller.ProvisionOptions) (*v1.PersistentVolume, controller.ProvisioningState, error) {
	parameters, err := NewStorageClassParameters(options.StorageClass.Parameters)
	if err != nil {
		return nil, controller.ProvisioningNoChange, fmt.Errorf("failed to parse StorageClass parameters: %w", err)
	}

	datasetPath := fmt.Sprintf("%s/%s", p.ParentDataset, options.PVName)
	properties := make(map[string]string)

	if !slices.Contains(options.PVC.Spec.AccessModes, v1.ReadWriteOnce) &&
		!slices.Contains(options.PVC.Spec.AccessModes, v1.ReadWriteOncePod) {
		return nil, controller.ProvisioningFinished, fmt.Errorf("ReadOnlyMany or ReadWriteMany access modes are not supported by this provisioner")
	}

	var reclaimPolicy v1.PersistentVolumeReclaimPolicy
	if options.StorageClass.ReclaimPolicy == nil {
		// Default is delete, see https://kubernetes.io/docs/concepts/storage/storage-classes/#reclaim-policy
		reclaimPolicy = v1.PersistentVolumeReclaimDelete
	} else if *options.StorageClass.ReclaimPolicy == v1.PersistentVolumeReclaimRecycle {
		return nil, controller.ProvisioningFinished, fmt.Errorf("unsupported reclaim policy of this provisioner: %s", v1.PersistentVolumeReclaimRecycle)
	} else {
		reclaimPolicy = *options.StorageClass.ReclaimPolicy
	}

	storageRequest := options.PVC.Spec.Resources.Requests[v1.ResourceStorage]
	storageRequestBytes := strconv.FormatInt(storageRequest.Value(), 10)
	properties[RefQuotaProperty] = storageRequestBytes
	properties[ManagedByProperty] = p.InstanceName
	properties[ReclaimPolicyProperty] = string(reclaimPolicy)

	if parameters.ReserveSpace {
		properties[RefReservationProperty] = storageRequestBytes
	}

	dataset, err := p.zfs.CreateDataset(datasetPath, properties)
	if err != nil {
		return nil, controller.ProvisioningFinished, fmt.Errorf("creating ZFS dataset failed: %w", err)
	}
	if err := p.zfs.SetPermissions(dataset); err != nil {
		return nil, controller.ProvisioningFinished, err
	}
	p.log.Info("dataset created", "dataset", dataset.Name)

	// Copy the annotations from the claim and overwrite with ours
	if options.PVC.Annotations == nil {
		options.PVC.Annotations = make(map[string]string)
	}
	annotations := options.PVC.Annotations
	annotations[DatasetPathAnnotation] = dataset.Name

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        options.PVName,
			Labels:      options.PVC.Labels,
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: reclaimPolicy,
			AccessModes:                   createAccessModes(options),
			Capacity: v1.ResourceList{
				v1.ResourceStorage: options.PVC.Spec.Resources.Requests[v1.ResourceStorage],
			},
			PersistentVolumeSource: createVolumeSource(parameters, dataset),
			NodeAffinity:           createNodeAffinity(parameters),
		},
	}
	return pv, controller.ProvisioningFinished, nil
}

func createAccessModes(options controller.ProvisionOptions) []v1.PersistentVolumeAccessMode {
	if slices.Contains(options.PVC.Spec.AccessModes, v1.ReadWriteOncePod) {
		return []v1.PersistentVolumeAccessMode{v1.ReadWriteOncePod}
	}
	return []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}
}

func createVolumeSource(parameters *ZFSStorageClassParameters, dataset *zfs.Dataset) v1.PersistentVolumeSource {
	hostPathType := v1.HostPathDirectory
	return v1.PersistentVolumeSource{
		HostPath: &v1.HostPathVolumeSource{
			Path: dataset.Mountpoint,
			Type: &hostPathType,
		},
	}
}

func createNodeAffinity(parameters *ZFSStorageClassParameters) *v1.VolumeNodeAffinity {
	node, _ := os.LookupEnv("ZFS_NODE_NAME")
	return &v1.VolumeNodeAffinity{Required: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{
		{
			MatchExpressions: []v1.NodeSelectorRequirement{
				{
					Values:   []string{node},
					Operator: v1.NodeSelectorOpIn,
					Key:      v1.LabelHostname,
				},
			},
		},
	}}}
}
