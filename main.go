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

func (opt *Opt) Validate(_ []string) error {
	if opt.IgnoreInterfaces != "" {
		var err error
		opt.ignoreInterfacesRegexp, err = regexp.Compile(opt.IgnoreInterfaces)
		if err != nil {
			return fmt.Errorf("invalid ignore-interfaces regexp: %w", err)
		}
	}
	return nil
}

func (opt *Opt) Run(_ []string) {
	u := LinuxNetDevPlugin{
		ignoreInterfaces:       opt.IgnoreInterfaces,
		ignoreInterfacesRegexp: opt.ignoreInterfacesRegexp,
		workDir:                pluginutil.PluginWorkDir(),
	}
	plugin := mp.NewMackerelPlugin(u)
	plugin.Run()
}

func main() {
	opt := &Opt{}
	os.Exit(flagrun.Ship(opt, flagrun.Version(version), flagrun.Validator(opt.Validate)))
}
