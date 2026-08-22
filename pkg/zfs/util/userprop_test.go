package zfs

import (
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidateUserProperty(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  codes.Code
	}{
		{"io.openebs.test:hello", "foo", codes.OK},
		{"io-openebs-test:hello", "foo", codes.OK},
		{"io_openebs_test:hello", "foo", codes.OK},
		{"io.openebs.test:hello:world", "foo", codes.OK},
		{"io.openebs.test::hello", "foo", codes.OK},
		{"io.openebs.test:", "foo", codes.OK},
		{"foo", "missing colon", codes.InvalidArgument},
		{"", "empty name", codes.InvalidArgument},
		{"-foo", "dash prefix name", codes.InvalidArgument},
		{"あ", "invalid bytes in name", codes.InvalidArgument},
		{"foo:bar", "あ", codes.OK},
		{"foo:stringtoolong", strings.Repeat("a", 8193), codes.InvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserProperty(tt.name, tt.value)
			if got := status.Code(err); got != tt.want {
				if err != nil {
					t.Errorf("ValidateUserProperty(\"%v\", \"%v\") = %v (%v), want %v", tt.name, tt.value, got, status.Convert(err).Message(), tt.want)
				} else {
					t.Errorf("ValidateUserProperty(\"%v\", \"%v\") = %v, want %v", tt.name, tt.value, got, tt.want)
				}
			}
		})
	}
}
