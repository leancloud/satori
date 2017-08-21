package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/leancloud/satori/agent/cgroups"
	"github.com/leancloud/satori/agent/cron"
	"github.com/leancloud/satori/agent/funcs"
	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/agent/http"
)

func main() {

	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	check := flag.Bool("check", false, "check collector")

	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}

	if *check {
		funcs.CheckCollector()
		os.Exit(0)
	}

	g.ParseConfig(*cfg)

	funcs.BuildMappers()

	cg := g.Config().Cgroups
	if cg != nil {
		if err := cgroups.JailMe("satori", cg.CPU, cg.Memory); err != nil {
			fmt.Println("Can't setup cgroups:", err)
			if cg.Panic {
				panic(err)
			}
		}
	}

	go cron.SyncWithMaster()
	go cron.StartCollect()
	go g.SendToTransferProc()
	go http.Start()

	select {}
}
