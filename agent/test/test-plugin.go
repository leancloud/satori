package main

import (
	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/agent/plugins"
	"github.com/leancloud/satori/common/model"
	"time"
)

//  func RunPlugins(dirs []string, metrics []model.PluginParam) {

func main() {
	g.ParseConfig("cfg.example.json")
	cfg := g.Config()
	cfg.Debug = true
	cfg.Plugin.Enabled = true
	cfg.Plugin.Subdir = "."
	cfg.Plugin.CheckoutPath = "."
	for {
		plugins.RunPlugins([]string{"test"}, []model.PluginParam{})
		time.Sleep(time.Second * 60)
	}
}
