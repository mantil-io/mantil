package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mantil-io/mantil/cli/cmd"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
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
	d := domain.Deployment()
	// if env is set prepare variables for usage in scripts/deploy.sh and exit
	// should be used as:
	//    eval $(MANTIL_ENV=1 mantil)
	//    ...
	//    # use $BUCKET in script
	// if $BUCKET2 is set upload to that location also
	// if $RELEASE is set this is release, not development, version
	fmt.Printf("export BUCKET='%s'\n", d.PutPath())
	if lb := d.LatestBucket(); lb != "" {
		fmt.Printf("export BUCKET2='%s'\n", lb)
	}
	if d.Release() {
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
		ui.Errorf("failed to open log file: %s", err)
	}
	defer log.Close()
	log.Printf("version: %s, args: %v", domain.Version(), domain.RemoveAWSCredentials(os.Args))
	if !isWatch() { // watch has it's own ctrlc handling
		go catchInterupt()
	}
	if err := cmd.Execute(); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func isWatch() bool {
	return len(os.Args) > 1 && os.Args[1] == "watch"
}

func catchInterupt() {
	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, syscall.SIGINT)
	log.Signal((<-ctrlc).String())
	log.Close()
	os.Exit(1)
}
