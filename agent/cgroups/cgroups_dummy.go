// +build !linux

package cgroups

import "fmt"

func JailMe(cgPath string, cpu float32, mem int64) error {
	return fmt.Errorf("Not on linux, cgroups not supported")
}
