package plugins

import (
	"log"

	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/common/model"
)

func SyncConfig(pluginVer string, pluginDirs []string, plugins []model.PluginParam) {
	debug := g.Config().Debug

	if pluginVer != "" {
		v, _ := GetCurrentPluginVersion()
		if pluginVer != v {
			log.Printf("Plugin version old[%s] != new[%s], update.", v, pluginVer)
			err := UpdatePlugin(pluginVer)
			if err == nil {
				err := TryUpdateAgent()
				log.Printf("TryUpdateAgent: %s\n", err)
			}
		}
	}

	if pluginDirs != nil || plugins != nil {
		if debug {
			log.Printf("Got PluginDirs: %s", pluginDirs)
			log.Printf("Got Plugins: %s", plugins)
		}
		RunPlugins(pluginDirs, plugins)
	}
}
