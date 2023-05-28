module github.com/openebs/zfs-localpv

go 1.19

require (
	github.com/container-storage-interface/spec v1.8.0
	github.com/docker/go-units v0.4.0
	github.com/golang/protobuf v1.5.3
	github.com/jpillora/go-ogle-analytics v0.0.0-20161213085824-14b04e0594ef
	github.com/kubernetes-csi/csi-lib-utils v0.6.1
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.27.7
	github.com/openebs/lib-csi v0.7.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.6.0
	github.com/stretchr/testify v1.8.1
	golang.org/x/net v0.10.0
	golang.org/x/sys v0.8.0
	google.golang.org/grpc v1.55.0
	k8s.io/api v0.27.2
	k8s.io/apimachinery v0.27.2
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.27.2
	k8s.io/klog/v2 v2.100.1
	k8s.io/utils v0.0.0-20230505201702-9f6742963106
	sigs.k8s.io/controller-runtime v0.2.0
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.10.2 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/imdario/mergo v0.3.15 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/oauth2 v0.8.0 // indirect
	golang.org/x/term v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.9.1 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/gengo v0.0.0-20220902162205-c0856e24416d // indirect
	k8s.io/kube-openapi v0.0.0-20230525220651-2546d827e515 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

replace (
	github.com/container-storage-interface/spec => github.com/container-storage-interface/spec v1.1.0
	k8s.io/api => k8s.io/api v0.27.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.27.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.27.2
	k8s.io/apiserver => k8s.io/apiserver v0.27.2
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.27.2
	k8s.io/client-go => k8s.io/client-go v0.27.2
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.27.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.27.2
	k8s.io/code-generator => k8s.io/code-generator v0.27.2
	k8s.io/component-base => k8s.io/component-base v0.27.2
	k8s.io/cri-api => k8s.io/cri-api v0.27.2
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.27.2
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.27.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.27.2
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.27.2
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.27.2
	k8s.io/kubectl => k8s.io/kubectl v0.27.2
	k8s.io/kubelet => k8s.io/kubelet v0.27.2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.27.2
	k8s.io/metrics => k8s.io/metrics v0.27.2
	k8s.io/node-api => k8s.io/node-api v0.27.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.27.2
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.27.2
	k8s.io/sample-controller => k8s.io/sample-controller v0.27.2
)
