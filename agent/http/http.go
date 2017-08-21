package http

import (
	"encoding/json"
	"github.com/leancloud/satori/agent/g"
	"log"
	"net/http"
	_ "net/http/pprof"
)

type Dto struct {
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func init() {
	http.HandleFunc("/v1/ping", httpPing)
	http.HandleFunc("/v1/push", httpPush)
	http.HandleFunc("/v1/plugin/reset", httpPluginReset)
}

func RenderJson(w http.ResponseWriter, v interface{}) {
	bs, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(bs)
}

func RenderDataJson(w http.ResponseWriter, data interface{}) {
	RenderJson(w, Dto{Msg: "success", Data: data})
}

func RenderMsgJson(w http.ResponseWriter, msg string) {
	RenderJson(w, map[string]string{"msg": msg})
}

func AutoRender(w http.ResponseWriter, data interface{}, err error) {
	if err != nil {
		RenderMsgJson(w, err.Error())
		return
	}

	RenderDataJson(w, data)
}

func Start() {
	addr := g.Config().HTTP
	if addr == "" {
		return
	}

	s := &http.Server{
		Addr:           addr,
		MaxHeaderBytes: 1 << 30,
	}

	log.Println("listening", addr)
	log.Fatalln(s.ListenAndServe())
}
