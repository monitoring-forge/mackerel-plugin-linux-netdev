package main

import (
	"fmt"
	"os"
	"regexp"

	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/prometheus/procfs"
)

var tooOldDuration = 600.0 // seconds

type LinuxNetDevPlugin struct {
	ignoreInterfaces       string
	ignoreInterfacesRegexp *regexp.Regexp
	workDir                string
}

func (u LinuxNetDevPlugin) GraphDefinition() map[string]mp.Graphs {
	graphdef := func(label string, metrics ...mp.Metrics) mp.Graphs {
		return mp.Graphs{
			Label:   label,
			Unit:    mp.UnitInteger,
			Metrics: metrics,
		}
	}
	metricsdef := func(name, label string) mp.Metrics {
		return mp.Metrics{
			Name:    name,
			Label:   label,
			Stacked: false,
		}
	}
	return map[string]mp.Graphs{
		"linux-netdev.errors.#": graphdef("Linux NetDev errors per sec",
			metricsdef("tx", "transmit errors encountered"),
			metricsdef("rx", "receive errors encountered"),
		),
		"linux-netdev.dropped.#": graphdef("Linux NetDev dropped packets per sec",
			metricsdef("tx", "packets dropped while transmitting"),
			metricsdef("rx", "packets dropped while receiving"),
		),
		"linux-netdev.pps.#": graphdef("Linux NetDev packets per sec",
			metricsdef("tx", "packets transmitted"),
			metricsdef("rx", "packets received"),
		),
	}
}

func (u LinuxNetDevPlugin) interfaceMetrics() (map[string]procfs.NetDevLine, error) {
	return u.interfaceMetricsFS("/proc")
}

func (u LinuxNetDevPlugin) interfaceMetricsFS(mountPoint string) (map[string]procfs.NetDevLine, error) {
	res := map[string]procfs.NetDevLine{}
	pfs, err := procfs.NewFS(mountPoint)
	if err != nil {
		return res, err
	}
	netdev, err := pfs.NetDev()
	if err != nil {
		return res, err
	}
	cur := map[string]procfs.NetDevLine{}
	for _, i := range netdev {
		if i.Name == "lo" {
			continue
		}
		if u.ignoreInterfaces != "" && u.ignoreInterfacesRegexp.MatchString(i.Name) {
			continue
		}
		cur[i.Name] = i
	}
	return cur, nil

}

func calcdiff(cur, prev uint64) float64 {
	if cur < prev {
		return 0
	}
	return float64(cur - prev)
}

func (u LinuxNetDevPlugin) calcMetrics(timeDiff float64, curMetrics, prevMetrics map[string]procfs.NetDevLine) map[string]float64 {
	res := map[string]float64{}
	allTxErrors := 0.0
	allTxDropped := 0.0
	allRxErrors := 0.0
	allRxDropped := 0.0
	for k, c := range curMetrics {
		p, ok := prevMetrics[k]
		if !ok {
			continue
		}
		// tx_errors
		txErrors := calcdiff(c.TxErrors, p.TxErrors)
		res[fmt.Sprintf("linux-netdev.errors.%s.tx", c.Name)] = txErrors / timeDiff
		allTxErrors += txErrors
		// tx_dropped
		txDropped := calcdiff(c.TxDropped, p.TxDropped)
		res[fmt.Sprintf("linux-netdev.dropped.%s.tx", c.Name)] = txDropped / timeDiff
		allTxDropped += txDropped
		// rx_errors
		rxErrors := calcdiff(c.RxErrors, p.RxErrors)
		res[fmt.Sprintf("linux-netdev.errors.%s.rx", c.Name)] = rxErrors / timeDiff
		allRxErrors += rxErrors
		// rx_dropped
		rxDropped := calcdiff(c.RxDropped, p.RxDropped)
		res[fmt.Sprintf("linux-netdev.dropped.%s.rx", c.Name)] = rxDropped / timeDiff
		allRxDropped += rxDropped
		// tx_packets
		txPackets := calcdiff(c.TxPackets, p.TxPackets)
		res[fmt.Sprintf("linux-netdev.pps.%s.tx", c.Name)] = txPackets / timeDiff
		// tx_packets
		rxPackets := calcdiff(c.RxPackets, p.RxPackets)
		res[fmt.Sprintf("linux-netdev.pps.%s.rx", c.Name)] = rxPackets / timeDiff
	}

	res["linux-netdev.errors.all.tx"] = allTxErrors / timeDiff
	res["linux-netdev.dropped.all.tx"] = allTxDropped / timeDiff
	res["linux-netdev.errors.all.rx"] = allRxErrors / timeDiff
	res["linux-netdev.dropped.all.rx"] = allRxDropped / timeDiff

	return res
}

func (u LinuxNetDevPlugin) FetchMetrics() (map[string]float64, error) {
	curMetrics, err := u.interfaceMetrics()
	if err != nil {
		return map[string]float64{}, err
	}

	path := statFile()

	defer func() {
		err := writeStats(u.workDir, path, curMetrics)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}()

	if !fileExists(u.workDir, path) {
		fmt.Fprintf(os.Stderr, "Notice: first time execution command\n")
		return map[string]float64{}, nil
	}

	timeDiff, prevMetrics, err := readStats(u.workDir, path)
	if err != nil {
		return map[string]float64{}, err
	}
	res := u.calcMetrics(timeDiff, curMetrics, prevMetrics)

	return res, nil
}
