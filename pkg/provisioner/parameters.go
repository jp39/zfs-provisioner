package provisioner

import (
	"fmt"
	"strings"
)

const (
	ReserveSpaceParameter = "reserveSpace"
)

// StorageClass Parameters are expected in the following schema:
/*
parameters:
  reserveSpace: true|false
*/

type (
	// ZFSStorageClassParameters represents the parameters on the `StorageClass`
	// object. It is used to ease access and validate those parameters at run time.
	ZFSStorageClassParameters struct {
		ReserveSpace bool
	}
)

// NewStorageClassParameters takes a storage class parameters, validates it for invalid configuration and returns a
// ZFSStorageClassParameters on success.
func NewStorageClassParameters(parameters map[string]string) (*ZFSStorageClassParameters, error) {
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
		ReserveSpace: reserveSpace,
	}

	return p, nil
}
