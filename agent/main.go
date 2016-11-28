package main

import (
	"flag"
	"fmt"
	"os"

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

	go cron.SyncWithMaster()
	go cron.StartCollect()
	go http.Start()

	select {}

}
