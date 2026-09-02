package zfs

import (
	"fmt"
	"strings"
)

// fsTypeDefault is the filesystem an empty fstype stands for. A volume
// capability is allowed to leave the type out, a statically provisioned volume
// often does, and k8s.io/mount-utils formats such a volume as ext4.
const fsTypeDefault = "ext4"

// formattableFsTypes are the filesystems the driver puts on a zvol, the only
// ones a mkfs option can be meant for. A "zfs" volume is a dataset and is never
// formatted.
var formattableFsTypes = map[string]bool{
	"ext2":  true,
	"ext3":  true,
	"ext4":  true,
	"xfs":   true,
	"btrfs": true,
}

// defaultFormatOptions holds the mkfs options a node uses for a filesystem when
// the storage class of the volume does not ask for any. It is set once, from the
// --default-format-options flags of the node agent.
var defaultFormatOptions = map[string][]string{}

// SetDefaultFormatOptions takes the "<fstype>=<options>" entries of the node
// agent and keeps them as the format defaults of this node.
func SetDefaultFormatOptions(entries []string) error {
	options := make(map[string][]string, len(entries))

	for _, entry := range entries {
		fstype, value, found := strings.Cut(entry, "=")
		if !found {
			return fmt.Errorf("format options %q are not in the <fstype>=<options> form", entry)
		}

		fstype = strings.ToLower(strings.TrimSpace(fstype))
		if !formattableFsTypes[fstype] {
			return fmt.Errorf("format options %q are for %q, which is not one of the formatted filesystems", entry, fstype)
		}

		if value := strings.Fields(value); len(value) > 0 {
			options[fstype] = value
		}
	}

	defaultFormatOptions = options

	return nil
}

// FormatOptions returns the extra mkfs options for the given filesystem, the
// ones asked for by the storage class if there are any, the defaults of this
// node otherwise.
func FormatOptions(fstype, scOptions string) []string {
	if options := strings.Fields(scOptions); len(options) > 0 {
		return options
	}

	fstype = strings.ToLower(strings.TrimSpace(fstype))
	if fstype == "" {
		fstype = fsTypeDefault
	}

	return defaultFormatOptions[fstype]
}
