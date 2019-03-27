package http

import (
	"encoding/json"
	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/common/model"
	"net/http"
)

func httpPush(w http.ResponseWriter, req *http.Request) {
	if req.ContentLength == 0 {
		http.Error(w, "body is blank", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(req.Body)
	var metrics []*model.MetricValue
	err := decoder.Decode(&metrics)
	if err != nil {
		http.Error(w, "cannot decode body", http.StatusBadRequest)
		return
	}

	g.SendToTransfer(metrics)
	w.Write([]byte("success"))
}
