// +build !windows

package main

import (
	"syscall"
)

func SDOSChroot(directory string) {
	syscall.Chroot(directory)
}
