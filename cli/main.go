package main

import (
	"fmt"
	"os"

	"github.com/mantil-io/mantil/cli/build"
	"github.com/mantil-io/mantil/cli/cmd"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/event"
	"github.com/mantil-io/mantil/event/net"
)

const (
	showEnv   = "MANTIL_ENV"
	genDocEnv = "MANTIL_GEN_DOC"
)

func main() {
	if printEnv() {
		return
	}
	if genDoc() {
		return
	}
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func printEnv() (ok bool) {
	if _, ok = os.LookupEnv(showEnv); !ok {
		return
	}
	v := build.Version() //    setup.NewVersion(tag, dev, ontag)
	// if env is set prepare variables for usage in scripts/deploy.sh and exit
	// should be used as:
	//    eval $(MANTIL_ENV=1 mantil)
	//    ...
	//    # use $BUCKET in script
	// if $BUCKET2 is set upload to that location also
	// if $RELEASE is set this is release, not development, version
	fmt.Printf("export BUCKET='%s'\n", v.UploadBucket())
	if lb := v.LatestBucket(); lb != "" {
		fmt.Printf("export BUCKET2='%s'\n", lb)
	}
	if v.Release() {
		fmt.Printf("export RELEASE='1'\n")
	}
	return
}

func genDoc() (ok bool) {
	var dir string
	if dir, ok = os.LookupEnv(genDocEnv); !ok {
		return
	}
	if err := cmd.GenDoc(dir); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	return
}

func run() error {
	if err := log.Open(); err != nil {
		ui.Errorf("failed to open log file: %s\n", err)
	}
	defer log.Close()
	v := build.Version()

	var cc event.CliCommand
	cc.Start()
	defer func() {
		cc.End()
		p, err := net.NewPublisher()
		if err != nil {
			log.Error(err)
			return
		}
		buf, err := cc.Marshal()
		if err != nil {
			log.Error(err)
			return
		}
		if err := p.Pub(buf); err != nil {
			log.Error(err)
			return
		}
	}()

	cc.Version = v.String()
	cc.Args = os.Args
	// TODO add other attributes and events

	log.Printf("build time data:: %s", build.Log())
	log.Printf("version: %s, bucket: %s", v.String(), v.UploadBucket())
	log.Printf("args: %v", os.Args)

	if err := cmd.Execute(); err != nil {
		cc.Error = err.Error()
		// TODO add call stack also
		log.Error(err)
		return err
	}
	return nil
}
