module github.com/openebs/zfs-localpv

go 1.14

require (
	cloud.google.com/go v0.49.0 // indirect
	github.com/container-storage-interface/spec v1.1.0
	github.com/docker/go-units v0.4.0
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/jpillora/go-ogle-analytics v0.0.0-20161213085824-14b04e0594ef
	github.com/kubernetes-csi/csi-lib-utils v0.6.1
	github.com/onsi/ginkgo v1.10.3
	github.com/onsi/gomega v1.7.1
	github.com/openebs/lib-csi v0.0.0-20201218144414-a64be5d4731e
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	golang.org/x/sys v0.0.0-20190902133755-9109b7679e13
	google.golang.org/grpc v1.23.1
	k8s.io/api v0.15.12
	k8s.io/apimachinery v0.15.12
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.15.12
	k8s.io/gengo v0.0.0-20190826232639-a874a240740c // indirect
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a // indirect
	k8s.io/kubernetes v1.15.12
	sigs.k8s.io/controller-runtime v0.2.0
)

replace (
	k8s.io/api => k8s.io/api v0.15.12
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.15.12
	k8s.io/apimachinery => k8s.io/apimachinery v0.15.12
	k8s.io/apiserver => k8s.io/apiserver v0.15.12
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.15.12
	k8s.io/client-go => k8s.io/client-go v0.15.12
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.15.12
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.15.12
	k8s.io/code-generator => k8s.io/code-generator v0.15.12
	k8s.io/component-base => k8s.io/component-base v0.15.12
	k8s.io/cri-api => k8s.io/cri-api v0.15.12
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.15.12
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.15.12
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.15.12
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.15.12
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.15.12
	k8s.io/kubectl => k8s.io/kubectl v0.15.12
	k8s.io/kubelet => k8s.io/kubelet v0.15.12
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.15.12
	k8s.io/metrics => k8s.io/metrics v0.15.12
	k8s.io/node-api => k8s.io/node-api v0.15.12
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.15.12
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.15.12
	k8s.io/sample-controller => k8s.io/sample-controller v0.15.12
)
