package funcs

import (
	"github.com/leancloud/satori/common/model"
	"github.com/toolkits/nux"
	"log"
)

func KernelMetrics() (L []*model.MetricValue) {

	maxFiles, err := nux.KernelMaxFiles()
	if err != nil {
		log.Println(err)
		return
	}

	L = append(L, V("kernel.maxfiles", float64(maxFiles)))

	maxProc, err := nux.KernelMaxProc()
	if err != nil {
		log.Println(err)
		return
	}

	L = append(L, V("kernel.maxproc", float64(maxProc)))

	allocateFiles, err := nux.KernelAllocateFiles()
	if err != nil {
		log.Println(err)
		return
	}

	L = append(L, V("kernel.files.allocated", float64(allocateFiles)))
	L = append(L, V("kernel.files.left", float64(maxFiles-allocateFiles)))
	return
}
