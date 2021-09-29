package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mantil-io/mantil/cli/cmd"
	"github.com/mantil-io/mantil/cli/commands/setup"
)

var (
	tag   string
	dev   string
	ontag string
)

const showEnv = "MANTIL_ENV"

func main() {
	v := setup.NewVersion(tag, dev, ontag)

	if _, ok := os.LookupEnv(showEnv); ok {
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

	ctx := setup.SetVersion(context.Background(), v)
	root := cmd.Root()
	root.Version = v.String()
	if err := root.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
