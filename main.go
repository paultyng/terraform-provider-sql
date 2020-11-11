package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/hashicorp/go-plugin"
	server "github.com/hashicorp/terraform-plugin-go/tfprotov5/server"

	"github.com/paultyng/terraform-provider-sql/internal/provider"
)

// Generate docs for website
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "dev"

	// goreleaser can also pass the specific commit if you want
	// commit  string = ""
)

const (
	providerAddr = "registry.terraform.io/paultyng/sql"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	ctx := context.Background()
	opts := []server.ServeOpt{}
	close := make(chan struct{})

	if debugMode {
		reattach := make(chan *plugin.ReattachConfig)

		go func() {
			for {
				select {
				case conf := <-reattach:
					reattachBytes, err := json.Marshal(map[string]struct {
						Protocol string
						Pid      int
						Test     bool
						Addr     struct {
							Network string
							String  string
						}
					}{
						providerAddr: {
							Protocol: string(conf.Protocol),
							Pid:      conf.Pid,
							Test:     conf.Test,
							Addr: struct {
								Network string
								String  string
							}{
								Network: conf.Addr.Network(),
								String:  conf.Addr.String(),
							},
						},
					})
					if err != nil {
						panic(fmt.Sprintf("Error building reattach string: %s", err))
					}

					reattachStr := string(reattachBytes)

					fmt.Printf("Provider started, to attach Terraform set the TF_REATTACH_PROVIDERS env var:\n\n")
					switch runtime.GOOS {
					case "windows":
						fmt.Printf("\tCommand Prompt:\tset \"TF_REATTACH_PROVIDERS=%s\"\n", reattachStr)
						fmt.Printf("\tPowerShell:\t$env:TF_REATTACH_PROVIDERS='%s'\n", strings.ReplaceAll(reattachStr, `'`, `''`))
					case "linux", "darwin":
						fmt.Printf("\tTF_REATTACH_PROVIDERS='%s'\n", strings.ReplaceAll(reattachStr, `'`, `'"'"'`))
					default:
						fmt.Println(reattachStr)
					}
					fmt.Println("")
				case <-close:
				case <-ctx.Done():
				}
			}
		}()

		opts = append(opts, server.WithDebug(ctx, reattach, close))
	}

	var err error
	go func() {
		err = server.Serve(providerAddr, provider.New(version), opts...)
	}()

	select {
	case <-close:
	case <-ctx.Done():
	}
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
