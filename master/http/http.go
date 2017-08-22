package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/leancloud/satori/master/g"
	"github.com/leancloud/satori/master/state"
)

func addHandlers() {
	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		state.StateLock.RLock()
		s, err := json.Marshal(state.State)
		state.StateLock.RUnlock()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(s)
	})
}

func Start() {
	listen := g.Config().Http
	if listen == "" {
		return
	}

	addHandlers()

	s := &http.Server{
		Addr:           listen,
		MaxHeaderBytes: 1 << 30,
	}

	log.Println("starting REST API on", listen)
	log.Fatalln(s.ListenAndServe())
}
