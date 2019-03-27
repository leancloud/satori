package g

import (
	"log"
	"runtime"
)

// change log:
// 1.0.7: code refactor for open source
// 1.0.8: bugfix loop init cache
// 1.0.9: update host table anyway
// 1.1.0: remove Checksum when query plugins
// 2.0.0: Complete rewrite, remove almost everything
// 2.0.1: Remove state file, add REST API for retrieveing state
const (
	VERSION = "2.0.1"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
