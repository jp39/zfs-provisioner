package provisioner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStorageClassParameters(t *testing.T) {
	type args struct {
		parameters map[string]string
	}
	tests := []struct {
		name        string
		args        args
		want        *ZFSStorageClassParameters
		errContains string
	}{
		{
			name: "GivenWrongSpec_WhenTypeInvalid_ThenThrowError",
			args: args{
				parameters: map[string]string{
					TypeParameter: "invalid",
				},
			},
			errContains: TypeParameter,
		},
		{
			name: "GivenCorrectSpec_WhenTypeNfs_ThenReturnNfsParameters",
			args: args{
				parameters: map[string]string{
					TypeParameter:            "nfs",
					SharePropertiesParameter: "rw",
				},
			},
			want: &ZFSStorageClassParameters{NFSShareProperties: "rw"},
		},
		{
			name: "GivenCorrectSpec_WhenTypeNfsWithoutProperties_ThenReturnNfsParametersWithDefault",
			args: args{
				parameters: map[string]string{
					TypeParameter: "nfs",
				},
			},
			want: &ZFSStorageClassParameters{NFSShareProperties: "on"},
		},
		{
			name: "GivenCorrectSpec_WhenTypeHostPath_ThenReturnHostPathParameters",
			args: args{
				parameters: map[string]string{
					TypeParameter: "hostpath",
				},
			},
			want: &ZFSStorageClassParameters{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewStorageClassParameters(tt.args.parameters)
			if tt.errContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want.NFSShareProperties, result.NFSShareProperties)
		})
	}
}
