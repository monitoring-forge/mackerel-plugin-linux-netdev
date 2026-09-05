module github.com/monitoring-forge/mackerel-plugin-linux-netdev

go 1.25.0

require (
	github.com/jessevdk/go-flags v1.6.1
	github.com/mackerelio/go-mackerel-plugin v0.1.6
	github.com/mackerelio/golib v1.2.2
	github.com/prometheus/procfs v0.22.0
)

require go.yaml.in/yaml/v3 v3.0.5 // indirect

require (
	github.com/stretchr/testify v1.12.1
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
)
