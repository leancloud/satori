package main

import (
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"github.com/leancloud/satori/swcollector/config"
	"github.com/leancloud/satori/swcollector/funcs"
	"github.com/leancloud/satori/swcollector/logging"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	buf, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		logging.Fatalln("Can't read config from stdin:", err)
	}

	config.ParseConfig(buf)
	cfg := config.Config()

	if cfg.LogFile != "" {
		logging.SetOutputFilename(cfg.LogFile)
	}

	funcs.StartCollect()
	log.Printf("swcollector started, config %+v\n", cfg)

	select {}
}
