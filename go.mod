module github.com/buildtool/build-tools

go 1.16

require (
	cloud.google.com/go v0.81.0 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Microsoft/hcsshim v0.8.15 // indirect
	github.com/aws/aws-sdk-go v1.38.13
	github.com/caarlos0/env/v6 v6.5.0
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/containerd/cgroups v0.0.0-20210114181951-8a68de567b68 // indirect
	github.com/containerd/continuity v0.0.0-20210208174643-50096c924a4e // indirect
	github.com/daviddengcn/go-colortext v1.0.0 // indirect
	github.com/docker/docker v20.10.5+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/go-git/go-git/v5 v5.3.0
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.12
	github.com/liamg/tml v0.4.0
	github.com/mattn/go-colorable v0.1.8
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	gotest.tools/v3 v3.0.3 // indirect
	k8s.io/client-go v0.20.5
	k8s.io/kubectl v0.20.5
)

//replace github.com/docker/spdystream => github.com/moby/spdystream  v0.2.0
