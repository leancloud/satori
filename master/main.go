package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/leancloud/satori/master/g"
	"github.com/leancloud/satori/master/http"
	"github.com/leancloud/satori/master/rpc"
	"github.com/leancloud/satori/master/state"
)

func main() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}

	g.ParseConfig(*cfg)

	go rpc.Start()
	go state.Start()
	go http.Start()

	select {}
}
