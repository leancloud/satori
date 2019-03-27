package main

import (
	"flag"
	"fmt"
	"github.com/leancloud/satori/transfer/g"
	"github.com/leancloud/satori/transfer/receiver"
	"github.com/leancloud/satori/transfer/sender"
	"os"
)

func main() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	versionGit := flag.Bool("vg", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}
	if *versionGit {
		fmt.Println(g.VERSION, g.COMMIT)
		os.Exit(0)
	}

	// global config
	g.ParseConfig(*cfg)

	sender.Start()
	receiver.Start()

	select {}
}
