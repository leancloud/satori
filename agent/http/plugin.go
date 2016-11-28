package http

import (
	"net/http"

	"github.com/leancloud/satori/agent/plugins"
)

func configPluginRoutes() {
	/*
		http.HandleFunc("/plugin/update", func(w http.ResponseWriter, r *http.Request) {
			err := plugins.UpdatePlugin()
			if err != nil {
				w.Write([]byte(err.Error()))
			} else {
				w.Write([]byte("success"))
			}
		})
	// */

	http.HandleFunc("/plugin/reset", func(w http.ResponseWriter, r *http.Request) {
		err := plugins.ForceResetPlugin()
		if err != nil {
			w.Write([]byte(err.Error()))
		} else {
			w.Write([]byte("success"))
		}
	})
}
