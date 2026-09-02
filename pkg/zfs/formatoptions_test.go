package zfs

import (
	"reflect"
	"testing"
)

func TestSetDefaultFormatOptions(t *testing.T) {
	tests := map[string]struct {
		entries   []string
		expected  map[string][]string
		expectErr bool
	}{
		"one filesystem": {
			entries:  []string{"xfs=-i nrext64=0"},
			expected: map[string][]string{"xfs": {"-i", "nrext64=0"}},
		},
		"more than one filesystem": {
			entries: []string{"xfs=-i nrext64=0", "ext4=-m 0 -O ^orphan_file"},
			expected: map[string][]string{
				"xfs":  {"-i", "nrext64=0"},
				"ext4": {"-m", "0", "-O", "^orphan_file"},
			},
		},
		"fstype is case insensitive": {
			entries:  []string{"XFS=-i nrext64=0"},
			expected: map[string][]string{"xfs": {"-i", "nrext64=0"}},
		},
		"extra spaces are dropped": {
			entries:  []string{"xfs=  -i   nrext64=0  "},
			expected: map[string][]string{"xfs": {"-i", "nrext64=0"}},
		},
		"empty options mean no default": {
			entries:  []string{"xfs="},
			expected: map[string][]string{},
		},
		"nothing given": {
			entries:  nil,
			expected: map[string][]string{},
		},
		"missing fstype": {
			entries:   []string{"-i nrext64=0"},
			expectErr: true,
		},
		"no separator at all": {
			entries:   []string{"xfs -i nrext64"},
			expectErr: true,
		},
		"filesystem that is never formatted": {
			entries:   []string{"zfs=-i nrext64=0"},
			expectErr: true,
		},
		"unknown filesystem": {
			entries:   []string{"reiserfs=-q"},
			expectErr: true,
		},
		"empty fstype": {
			entries:   []string{"=-i nrext64=0"},
			expectErr: true,
		},
	}

	defer func() { defaultFormatOptions = map[string][]string{} }()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			defaultFormatOptions = map[string][]string{}

			err := SetDefaultFormatOptions(test.entries)
			if test.expectErr {
				if err == nil {
					t.Fatalf("expected an error for %v", test.entries)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for %v: %v", test.entries, err)
			}

			if !reflect.DeepEqual(defaultFormatOptions, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, defaultFormatOptions)
			}
		})
	}
}

func TestFormatOptions(t *testing.T) {
	defer func() { defaultFormatOptions = map[string][]string{} }()

	if err := SetDefaultFormatOptions([]string{"xfs=-i nrext64=0", "ext4=-m 0"}); err != nil {
		t.Fatalf("can not set the defaults: %v", err)
	}

	tests := map[string]struct {
		fstype   string
		sc       string
		expected []string
	}{
		"default of the filesystem": {
			fstype: "xfs", sc: "", expected: []string{"-i", "nrext64=0"},
		},
		"storage class replaces the default": {
			fstype: "xfs", sc: "-b size=2048", expected: []string{"-b", "size=2048"},
		},
		"default of another filesystem": {
			fstype: "ext4", sc: "", expected: []string{"-m", "0"},
		},
		"filesystem without a default": {
			fstype: "btrfs", sc: "", expected: nil,
		},
		"storage class on a filesystem without a default": {
			fstype: "btrfs", sc: "-M", expected: []string{"-M"},
		},
		"fstype is case insensitive": {
			fstype: "XFS", sc: "", expected: []string{"-i", "nrext64=0"},
		},
		"blank storage class value falls back to the default": {
			fstype: "xfs", sc: "   ", expected: []string{"-i", "nrext64=0"},
		},
		// an empty fstype in the volume capability is formatted as ext4 by
		// k8s.io/mount-utils, so it has to pick up the ext4 default
		"no fstype means ext4": {
			fstype: "", sc: "", expected: []string{"-m", "0"},
		},
		"padded fstype": {
			fstype: "  xfs  ", sc: "", expected: []string{"-i", "nrext64=0"},
		},
		"no fstype with a storage class value": {
			fstype: "", sc: "-b size=2048", expected: []string{"-b", "size=2048"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := FormatOptions(test.fstype, test.sc); !reflect.DeepEqual(got, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, got)
			}
		})
	}
}
