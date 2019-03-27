package funcs

import (
	"github.com/leancloud/satori/common/model"
	"github.com/toolkits/nux"
	"log"
)

func DeviceMetrics() (L []*model.MetricValue) {
	mountPoints, err := nux.ListMountPoint()

	if err != nil {
		log.Println(err)
		return
	}

	var diskTotal uint64 = 0
	var diskUsed uint64 = 0

	for idx := range mountPoints {
		var du *nux.DeviceUsage
		du, err = nux.BuildDeviceUsage(mountPoints[idx][0], mountPoints[idx][1], mountPoints[idx][2])
		if err != nil {
			log.Println(err)
			continue
		}

		diskTotal += du.BlocksAll
		diskUsed += du.BlocksUsed

		tags := map[string]string{
			"mount":  du.FsFile,
			"fstype": du.FsVfstype,
		}
		L = append(L, VT("df.bytes.total", float64(du.BlocksAll), tags))
		L = append(L, VT("df.bytes.used", float64(du.BlocksUsed), tags))
		L = append(L, VT("df.bytes.free", float64(du.BlocksFree), tags))
		L = append(L, VT("df.bytes.used.percent", du.BlocksUsedPercent, tags))
		L = append(L, VT("df.bytes.free.percent", du.BlocksFreePercent, tags))
		L = append(L, VT("df.inodes.total", float64(du.InodesAll), tags))
		L = append(L, VT("df.inodes.used", float64(du.InodesUsed), tags))
		L = append(L, VT("df.inodes.free", float64(du.InodesFree), tags))
		L = append(L, VT("df.inodes.used.percent", du.InodesUsedPercent, tags))
		L = append(L, VT("df.inodes.free.percent", du.InodesFreePercent, tags))

	}

	if len(L) > 0 && diskTotal > 0 {
		L = append(L, V("df.statistics.total", float64(diskTotal)))
		L = append(L, V("df.statistics.used", float64(diskUsed)))
		L = append(L, V("df.statistics.used.percent", float64(diskUsed)*100.0/float64(diskTotal)))
	}

	return
}
