package patcher

import (
	_ "embed"
	"sigs.k8s.io/yaml"
)

//go:embed zfsvolume-crd.yaml
var vol []byte

//go:embed zfssnapshot-crd.yaml
var snap []byte

//go:embed zfsrestore-crd.yaml
var res []byte

func NewZfsVolumesCrd() ([]byte, error) {
	return yaml.YAMLToJSON(vol)
}

func NewZfsSnapshotsCrd() ([]byte, error) {
	return yaml.YAMLToJSON(snap)
}

func NewZfsRestoresCrd() ([]byte, error) {
	return yaml.YAMLToJSON(res)
}
