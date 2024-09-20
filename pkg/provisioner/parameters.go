package provisioner

import (
	"fmt"
	"strings"
)

const (
	SharePropertiesParameter = "shareProperties"
	TypeParameter            = "type"
	ReserveSpaceParameter    = "reserveSpace"
)

// StorageClass Parameters are expected in the following schema:
/*
parameters:
  parentDataset: tank/volumes
  type: nfs|hostPath|auto
  shareProperties: rw=10.0.0.0/8,no_root_squash
  node: my-zfs-host
  reserveSpace: true|false
*/

type ProvisioningType string

const (
	Nfs      ProvisioningType = "nfs"
	HostPath ProvisioningType = "hostPath"
	Auto     ProvisioningType = "auto"
)

type (
	// ZFSStorageClassParameters represents the parameters on the `StorageClass`
	// object. It is used to ease access and validate those parameters at run time.
	ZFSStorageClassParameters struct {
		Type     ProvisioningType
		// NFSShareProperties specifies additional properties to pass to 'zfs create sharenfs=%s'.
		NFSShareProperties string
		ReserveSpace     bool
	}
)

// NewStorageClassParameters takes a storage class parameters, validates it for invalid configuration and returns a
// ZFSStorageClassParameters on success.
func NewStorageClassParameters(parameters map[string]string) (*ZFSStorageClassParameters, error) {
	for _, parameter := range []string{TypeParameter} {
		value := parameters[parameter]
		if value == "" {
			return nil, fmt.Errorf("undefined required parameter: %s", parameter)
		}
	}

	reserveSpaceValue, reserveSpaceValuePresent := parameters[ReserveSpaceParameter]
	var reserveSpace bool
	if !reserveSpaceValuePresent || strings.EqualFold(reserveSpaceValue, "true") {
		reserveSpace = true
	} else if strings.EqualFold(reserveSpaceValue, "false") {
		reserveSpace = false
	} else {
		return nil, fmt.Errorf("invalid '%s' parameter value: %s", ReserveSpaceParameter, parameters[ReserveSpaceParameter])
	}

	p := &ZFSStorageClassParameters{
		ReserveSpace:  reserveSpace,
	}
	typeParam := parameters[TypeParameter]
	switch typeParam {
	case "hostpath", "hostPath", "HostPath", "Hostpath", "HOSTPATH":
		p.Type = HostPath
	case "nfs", "Nfs", "NFS":
		p.Type = Nfs
	case "auto", "Auto", "AUTO":
		p.Type = Auto
	default:
		return nil, fmt.Errorf("invalid '%s' parameter value: %s", TypeParameter, typeParam)
	}

	if p.Type == Nfs || p.Type == Auto {
		shareProps := parameters[SharePropertiesParameter]
		if shareProps == "" {
			shareProps = "on"
		}
		p.NFSShareProperties = shareProps
	}

	return p, nil
}
