package http

import (
	"net/http"

	"github.com/leancloud/satori/agent/g"
)

func httpPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"result\":\"pong\",\"version\":\"" + g.VERSION + "\"}"))
}
