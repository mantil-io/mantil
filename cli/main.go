package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mantil-io/mantil/cli/cmd"
	"github.com/mantil-io/mantil/cli/cmd/setup"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
)

var (
	tag   string
	dev   string
	ontag string
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
	v := setup.NewVersion(tag, dev, ontag)
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

	v := setup.NewVersion(tag, dev, ontag)
	if err := cmd.GenDoc(v.String(), dir); err != nil {
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
	v := setup.NewVersion(tag, dev, ontag)

	log.Printf("version tag: %s, ontag: %s, dev: %s", tag, ontag, dev)
	log.Printf("args: %v", os.Args)

	ctx := setup.SetVersion(context.Background(), v)
	if err := cmd.Execute(ctx, v.String()); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
