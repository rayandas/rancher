module github.com/rancher/rancher/pkg/client

go 1.20

replace k8s.io/client-go => github.com/rancher/client-go v1.27.4-rancher1

require (
	github.com/rancher/norman v0.0.0-20230831160711-5de27f66385d
	k8s.io/apimachinery v0.27.4
)

require (
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rancher/wrangler v1.1.1-0.20230831050635-df1bd5aae9df // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/klog/v2 v2.90.1 // indirect
)
