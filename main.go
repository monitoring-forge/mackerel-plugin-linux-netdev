package main

import (
	"os"

	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/mackerelio/golib/pluginutil"
	"github.com/monitoring-forge/flagrun"
)

var version string

type Opt struct {
	Version          bool       `short:"v" long:"version" description:"Show version"`
	IgnoreInterfaces flagrun.Re `long:"ignore-interfaces" description:"Regexp for interfaces name to ignore"`
	workDir          string
}

func (opt *Opt) Validator(_ []string) error {
	if opt.workDir == "" {
		opt.workDir = pluginutil.PluginWorkDir()
	}
	return nil
}

func (opt *Opt) Run(_ []string) {
	plugin := mp.NewMackerelPlugin(opt)
	plugin.Run()
}

func main() {
	opt := &Opt{}
	os.Exit(flagrun.Ship(opt, flagrun.Version(version), flagrun.Validator(opt.Validator)))
}
