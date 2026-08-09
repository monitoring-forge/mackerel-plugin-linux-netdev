package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/jessevdk/go-flags"
	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/mackerelio/golib/pluginutil"
)

var version string
var commit string

const (
	OK = iota
	WARNING
	CRITICAL
	UNKNOWN
)

type Opt struct {
	Version                bool   `short:"v" long:"version" description:"Show version"`
	IgnoreInterfaces       string `long:"ignore-interfaces" description:"Regexp for interfaces name to ignore"`
	ignoreInterfacesRegexp *regexp.Regexp
}

func main() {
	os.Exit(_main())
}

func _main() int {
	opt := &Opt{}
	psr := flags.NewParser(opt, flags.HelpFlag|flags.PassDoubleDash)
	_, err := psr.Parse()
	if opt.Version {
		if commit == "" {
			commit = "dev"
		}
		fmt.Printf(
			"%s-%s\n%s/%s, %s, %s\n",
			filepath.Base(os.Args[0]),
			version,
			runtime.GOOS,
			runtime.GOARCH,
			runtime.Version(),
			commit)
		return OK
	} else if flags.WroteHelp(err) {
		fmt.Fprintf(os.Stdout, "%v\n", err)
		return OK
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return UNKNOWN
	}

	if opt.IgnoreInterfaces != "" {
		opt.ignoreInterfacesRegexp = regexp.MustCompile(opt.IgnoreInterfaces)
	}

	u := LinuxNetDevPlugin{
		ignoreInterfaces:       opt.IgnoreInterfaces,
		ignoreInterfacesRegexp: opt.ignoreInterfacesRegexp,
		workDir:                pluginutil.PluginWorkDir(),
	}
	plugin := mp.NewMackerelPlugin(u)
	plugin.Run()
	return OK
}
