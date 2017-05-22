package http

import (
	"net/http"

	"github.com/leancloud/satori/agent/plugins"
)

func httpPluginReset(w http.ResponseWriter, r *http.Request) {
	err := plugins.ForceResetPlugin()
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("success"))
	}
}
