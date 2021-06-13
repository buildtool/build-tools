module github.com/buildtool/build-tools

go 1.16

require (
	cloud.google.com/go v0.81.0 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Microsoft/hcsshim v0.8.16 // indirect
	github.com/alecthomas/kong v0.2.16
	github.com/apex/log v1.9.0
	github.com/aws/aws-sdk-go v1.38.17
	github.com/caarlos0/env/v6 v6.5.0
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/containerd/console v1.0.1
	github.com/daviddengcn/go-colortext v1.0.0 // indirect
	github.com/docker/docker v20.10.5+incompatible
	github.com/go-git/go-git/v5 v5.3.0
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.12
	github.com/liamg/tml v0.4.0
	github.com/moby/buildkit v0.8.3
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/tonistiigi/fsutil v0.0.0-20201103201449-0834f99b7b85
	gitlab.com/unboundsoftware/apex-mocks v0.0.4
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/client-go v0.21.0
	k8s.io/kubectl v0.21.0
)

replace (
	// protobuf: corresponds to containerd
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.5
	github.com/hashicorp/go-immutable-radix => github.com/tonistiigi/go-immutable-radix v0.0.0-20170803185627-826af9ccf0fe
	github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305
	// genproto: corresponds to containerd
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200224152610-e50cd9704f63
	// grpc: corresponds to protobuf
	google.golang.org/grpc => google.golang.org/grpc v1.30.0
)
