package driver

import (
	"testing"

	zfsapi "github.com/openebs/zfs-localpv/pkg/apis/openebs.io/zfs/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBackupSnapshotIndexFunc(t *testing.T) {
	tests := map[string]struct {
		input          interface{}
		expectedKeys   []string
		expectedToFail bool
	}{
		"Valid backup with all fields": {
			input: &zfsapi.ZFSBackup{
				Spec: zfsapi.ZFSBackupSpec{
					VolumeName:  "test-volume",
					OwnerNodeID: "test-node",
					SnapName:    "test-snap",
				},
			},
			expectedKeys:   []string{"test-volume.test-node.test-snap"},
			expectedToFail: false,
		},
		"Non-backup input": {
			input:          "not-a-backup",
			expectedKeys:   nil,
			expectedToFail: false,
		},
		"Empty fields": {
			input: &zfsapi.ZFSBackup{
				Spec: zfsapi.ZFSBackupSpec{
					VolumeName:  "",
					OwnerNodeID: "",
					SnapName:    "",
				},
			},
			expectedKeys:   []string{".."},
			expectedToFail: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			keys, err := BackupSnapshotIndexFunc(test.input)
			if test.expectedToFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedKeys, keys)
			}
		})
	}
}

func TestBackupPreviousSnapshotIndexFunc(t *testing.T) {
	tests := map[string]struct {
		input          interface{}
		expectedKeys   []string
		expectedToFail bool
	}{
		"Valid backup with prev snapshot": {
			input: &zfsapi.ZFSBackup{
				Spec: zfsapi.ZFSBackupSpec{
					VolumeName:   "test-volume",
					OwnerNodeID:  "test-node",
					PrevSnapName: "prev-snap",
				},
			},
			expectedKeys:   []string{"test-volume.test-node.prev-snap"},
			expectedToFail: false,
		},
		"Backup without prev snapshot": {
			input: &zfsapi.ZFSBackup{
				Spec: zfsapi.ZFSBackupSpec{
					VolumeName:   "test-volume",
					OwnerNodeID:  "test-node",
					PrevSnapName: "",
				},
			},
			expectedKeys:   nil,
			expectedToFail: false,
		},
		"Non-backup input": {
			input:          "not-a-backup",
			expectedKeys:   nil,
			expectedToFail: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			keys, err := BackupPreviousSnapshotIndexFunc(test.input)
			if test.expectedToFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedKeys, keys)
			}
		})
	}
}

func TestGetBackupKeyBySpec(t *testing.T) {
	bgc := &BackupGarbageCollector{}

	tests := map[string]struct {
		volumeName  string
		ownerNodeID string
		snapName    string
		expected    string
	}{
		"All fields populated": {
			volumeName:  "test-volume",
			ownerNodeID: "test-node",
			snapName:    "test-snap",
			expected:    "test-volume.test-node.test-snap",
		},
		"Empty fields": {
			volumeName:  "",
			ownerNodeID: "",
			snapName:    "",
			expected:    "..",
		},
		"Special characters": {
			volumeName:  "vol-123",
			ownerNodeID: "node_456",
			snapName:    "snap-789",
			expected:    "vol-123.node_456.snap-789",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := bgc.getBackupKeyBySpec(test.volumeName, test.ownerNodeID, test.snapName)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestParseBackupsFromIndexResults(t *testing.T) {
	bgc := &BackupGarbageCollector{}

	// Create test backup objects
	backup1 := &zfsapi.ZFSBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name: "backup1",
		},
	}
	backup2 := &zfsapi.ZFSBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name: "backup2",
		},
	}

	tests := map[string]struct {
		input          []interface{}
		expectedCount  int
		expectedNames  []string
		expectedToFail bool
	}{
		"Valid backup objects": {
			input:          []interface{}{backup1, backup2},
			expectedCount:  2,
			expectedNames:  []string{"backup1", "backup2"},
			expectedToFail: false,
		},
		"Empty input": {
			input:          []interface{}{},
			expectedCount:  0,
			expectedNames:  []string{},
			expectedToFail: false,
		},
		"Mixed valid and invalid objects": {
			input:          []interface{}{backup1, "not-a-backup"},
			expectedCount:  0,
			expectedNames:  nil,
			expectedToFail: true,
		},
		"All invalid objects": {
			input:          []interface{}{"string1", 123},
			expectedCount:  0,
			expectedNames:  nil,
			expectedToFail: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := bgc.parseBackupsFromIndexResults(test.input)

			if test.expectedToFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedCount, len(result))

				if len(test.expectedNames) > 0 {
					names := make([]string, len(result))
					for i, backup := range result {
						names[i] = backup.Name
					}
					assert.Equal(t, test.expectedNames, names)
				}
			}
		})
	}
}
