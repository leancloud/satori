package plugins

import (
	"log"

	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/common/model"
)

func SyncConfig(pluginVer string, pluginDirs []string, pluginMetrics []model.PluginParam) {
	debug := g.Config().Debug

	if pluginVer != "" {
		v, _ := GetCurrentPluginVersion()
		if pluginVer != v {
			if debug {
				log.Printf("Plugin version old[%s] != new[%s], update.", v, pluginVer)
			}
			err := UpdatePlugin(pluginVer)
			if err == nil {
				err := TryUpdate()
				if debug {
					log.Printf("TryUpdate: %s\n", err)
				}
			}
		}
	}

	if pluginDirs != nil || pluginMetrics != nil {
		if debug {
			log.Printf("Got PluginDirs: %s", pluginDirs)
			log.Printf("Got PluginMetrics: %s", pluginMetrics)
		}
		RunPlugins(pluginDirs, pluginMetrics)
	}
}
