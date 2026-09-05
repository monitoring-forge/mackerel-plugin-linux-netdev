package main

import (
	"fmt"
	"os"
	"regexp"

	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/mackerelio/golib/pluginutil"
	"github.com/monitoring-forge/flagrun"
)

var version string

type Opt struct {
	Version                bool   `short:"v" long:"version" description:"Show version"`
	IgnoreInterfaces       string `long:"ignore-interfaces" description:"Regexp for interfaces name to ignore"`
	ignoreInterfacesRegexp *regexp.Regexp
}

func (opt *Opt) Run(_ []string) (any, int) {
	if opt.IgnoreInterfaces != "" {
		var err error
		opt.ignoreInterfacesRegexp, err = regexp.Compile(opt.IgnoreInterfaces)
		if err != nil {
			return fmt.Errorf("invalid ignore-interfaces regexp: %w", err), flagrun.UNKNOWN
		}
	}

	u := LinuxNetDevPlugin{
		ignoreInterfaces:       opt.IgnoreInterfaces,
		ignoreInterfacesRegexp: opt.ignoreInterfacesRegexp,
		workDir:                pluginutil.PluginWorkDir(),
	}
	plugin := mp.NewMackerelPlugin(u)
	plugin.Run()
	return "", flagrun.OK

}

func main() {
	os.Exit(flagrun.Go(&Opt{}, flagrun.Version(version)))
}
