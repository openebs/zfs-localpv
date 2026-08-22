package zfs

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValidateUserProperty checks that the user property name and value are valid,
// conforming to the requirements outlined in zfsprops(7).
func ValidateUserProperty(name string, value string) error {
	if len(name) == 0 {
		return status.Error(codes.InvalidArgument, "name cannot be empty")
	}
	if len(name) > 256 {
		return status.Error(codes.InvalidArgument, "name too long")
	}
	if name[0] == '-' {
		return status.Error(codes.InvalidArgument, "name cannot begin with '-'")
	}

	var nameContainsColon = false
	for _, b := range []byte(name) {
		switch {
		case b >= 'a' && b <= 'z':
		case b >= '0' && b <= '9':
		case b == '-' || b == '.' || b == '_':
		case b == ':':
			nameContainsColon = true
		default:
			return status.Error(codes.InvalidArgument, "name contains invalid byte(s)")
		}
	}
	if !nameContainsColon {
		return status.Error(codes.InvalidArgument, "name does not contain ':'")
	}

	if len(value) > 8192 {
		return status.Error(codes.InvalidArgument, "value too long")
	}
	return nil
}
