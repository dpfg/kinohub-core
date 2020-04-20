package main

import (
	"os"

	"github.com/dpfg/kinohub-core/cmd"
	"github.com/jessevdk/go-flags"
)

const (
	defaultPort = 8081

	zeroConfName    = "KinoHub"
	zeroConfService = "_kinohub._tcp"
	zeroConfDomain  = "local."
)

type Opts struct {
	ServerCmd cmd.ServerCommand `command:"server"`
}

func main() {

	var opts Opts
	p := flags.NewParser(&opts, flags.Default)

	if _, err := p.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}
