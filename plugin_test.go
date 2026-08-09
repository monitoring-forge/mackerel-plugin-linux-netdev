package main

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/prometheus/procfs"
)

// writeNetDev creates a fake /proc/net/dev file under the given procDir.
func writeNetDev(t *testing.T, procDir string, content string) {
	t.Helper()
	netDir := filepath.Join(procDir, "net")
	if err := os.MkdirAll(netDir, 0o755); err != nil {
		t.Fatalf("failed to create net dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(netDir, "dev"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write net/dev: %v", err)
	}
}

const sampleNetDev = `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo:    1000      10    0    0    0     0          0         0     1000      10    0    0    0     0       0          0
  eth0:   20000     200    2    1    0     0          0         0    30000     300    3    2    0     0       0          0
  eth1:   40000     400    4    3    0     0          0         0    50000     500    5    4    0     0       0          0
`

func TestInterfaceMetricsFS(t *testing.T) {
	procDir := t.TempDir()
	writeNetDev(t, procDir, sampleNetDev)

	p := LinuxNetDevPlugin{}
	got, err := p.interfaceMetricsFS(procDir)
	if err != nil {
		t.Fatalf("interfaceMetricsFS failed: %v", err)
	}

	if _, ok := got["lo"]; ok {
		t.Fatalf("loopback interface 'lo' should be excluded")
	}

	wantKeys := map[string]bool{
		"eth0": true,
		"eth1": true,
	}
	if len(got) != len(wantKeys) {
		t.Fatalf("unexpected number of interfaces, got=%d want=%d", len(got), len(wantKeys))
	}
	for name := range wantKeys {
		if _, ok := got[name]; !ok {
			t.Fatalf("missing interface: %s", name)
		}
	}

	if got["eth0"].RxPackets != 200 || got["eth0"].TxPackets != 300 {
		t.Fatalf("unexpected eth0 metrics: %+v", got["eth0"])
	}
	if got["eth1"].RxPackets != 400 || got["eth1"].TxPackets != 500 {
		t.Fatalf("unexpected eth1 metrics: %+v", got["eth1"])
	}
}

func TestInterfaceMetricsFSIgnoreInterfaces(t *testing.T) {
	procDir := t.TempDir()
	writeNetDev(t, procDir, sampleNetDev)

	re := regexp.MustCompile(`^eth0$`)
	p := LinuxNetDevPlugin{
		ignoreInterfaces:       "^eth0$",
		ignoreInterfacesRegexp: re,
	}
	got, err := p.interfaceMetricsFS(procDir)
	if err != nil {
		t.Fatalf("interfaceMetricsFS failed: %v", err)
	}

	if _, ok := got["eth0"]; ok {
		t.Fatalf("eth0 should be ignored")
	}
	if _, ok := got["eth1"]; !ok {
		t.Fatalf("eth1 should not be ignored")
	}
}

func TestInterfaceMetricsFSInvalidMountPoint(t *testing.T) {
	p := LinuxNetDevPlugin{}
	_, err := p.interfaceMetricsFS(filepath.Join(t.TempDir(), "nonexistent"))
	if err == nil {
		t.Fatalf("expected error for nonexistent mount point")
	}
}

func TestCalcdiff(t *testing.T) {
	cases := []struct {
		cur  uint64
		prev uint64
		want float64
	}{
		{cur: 10, prev: 5, want: 5},
		{cur: 0, prev: 0, want: 0},
	}

	for _, c := range cases {
		got := calcdiff(c.cur, c.prev)
		if got != c.want {
			t.Fatalf("calcdiff(%d, %d) = %v, want %v", c.cur, c.prev, got, c.want)
		}
	}
}

func TestCalcMetrics(t *testing.T) {
	cur := map[string]procfs.NetDevLine{
		"eth0": {
			Name:      "eth0",
			TxErrors:  10,
			RxErrors:  20,
			TxDropped: 30,
			RxDropped: 40,
			TxPackets: 100,
			RxPackets: 200,
		},
	}
	prev := map[string]procfs.NetDevLine{
		"eth0": {
			Name:      "eth0",
			TxErrors:  5,
			RxErrors:  10,
			TxDropped: 15,
			RxDropped: 20,
			TxPackets: 50,
			RxPackets: 100,
		},
	}

	p := LinuxNetDevPlugin{}
	res := p.calcMetrics(10, cur, prev)

	want := map[string]float64{
		"linux-netdev.errors.eth0.tx":  0.5,
		"linux-netdev.errors.eth0.rx":  1.0,
		"linux-netdev.dropped.eth0.tx": 1.5,
		"linux-netdev.dropped.eth0.rx": 2.0,
		"linux-netdev.pps.eth0.tx":     5.0,
		"linux-netdev.pps.eth0.rx":     10.0,
		"linux-netdev.errors.all.tx":   0.5,
		"linux-netdev.dropped.all.tx":  1.5,
		"linux-netdev.errors.all.rx":   1.0,
		"linux-netdev.dropped.all.rx":  2.0,
	}

	for k, v := range want {
		got, ok := res[k]
		if !ok {
			t.Fatalf("missing metric key: %s", k)
		}
		if got != v {
			t.Fatalf("%s: got %v, want %v", k, got, v)
		}
	}
}
