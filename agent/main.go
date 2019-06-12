package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/leancloud/satori/agent/cgroups"
	"github.com/leancloud/satori/agent/cron"
	"github.com/leancloud/satori/agent/funcs"
	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/agent/http"
)

func main() {

	cfg := flag.String("c", "agent-cfg.yaml", "configuration file")
	version := flag.Bool("v", false, "show version")

	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}

	g.ParseConfig(*cfg)
	funcs.BuildMappers()

	defer func() {
		if r := recover(); r != nil {
			g.LastMessage("main")
		}
	}()

	cg := g.Config().Cgroups
	if cg != nil {
		if err := cgroups.JailMe("satori", cg.CPU, cg.Memory); err != nil {
			log.Println("Can't setup cgroups:", err)
			if cg.Enforce {
				panic(err)
			}
		}
	}

	go cron.SyncWithMaster()
	go cron.StartCollect()
	go g.SendToTransferProc()
	go http.Start()
	go g.ReportLastMessage()

	select {}
}
