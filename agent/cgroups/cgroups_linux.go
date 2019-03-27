package cgroups

import (
	"os"

	cglib "github.com/containerd/cgroups"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func strippedHierarchy() ([]cglib.Subsystem, error) {
	ss, err := cglib.V1()
	if err != nil {
		return ss, err
	}

	rst := []cglib.Subsystem{}
	for _, i := range ss {
		s := i.Name()
		if s == "cpu" || s == "memory" {
			rst = append(rst, i)
		}
	}
	return rst, err
}

func JailMe(cgPath string, cpu float32, mem int64) error {
	mlimit := mem * 1024 * 1024
	cpuPeriod := uint64(1e5)
	cpuQuota := int64(float32(cpuPeriod) * cpu / 100.0)
	ctrl, err := cglib.New(strippedHierarchy, cglib.NestedPath(cgPath), &specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Quota:  &cpuQuota,
			Period: &cpuPeriod,
		},
		Memory: &specs.LinuxMemory{
			Limit: &mlimit,
		},
	})
	if err != nil {
		return err
	}
	err = ctrl.Add(cglib.Process{Pid: os.Getpid()})
	if err != nil {
		return err
	}
	return nil
}
