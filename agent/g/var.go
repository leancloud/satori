package g

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/kardianos/osext"
)

var (
	ConfigVersion int64
	BinaryHash    []byte
	BinaryPath    string
)

func init() {
	var err error
	BinaryPath, err = osext.Executable()
	if err != nil {
		panic(fmt.Errorf("Can't get binary path: %s", err.Error()))
	}

	h := sha256.New()
	self, err := os.Open(BinaryPath)
	if err != nil {
		panic(fmt.Errorf("Can't open self for read: %s", err.Error()))
	}
	if _, err := io.Copy(h, self); err != nil {
		panic(fmt.Errorf("Can't read contents of self: %s", err.Error()))
	}
	self.Close()
	BinaryHash = h.Sum(nil)
}
