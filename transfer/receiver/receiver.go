package receiver

import (
	"github.com/leancloud/satori/transfer/receiver/rpc"
)

func Start() {
	go rpc.StartRpc()
}
