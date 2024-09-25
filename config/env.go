package config

import "syscall"

func GetEnv(envStr string) string {
	val, ok := syscall.Getenv(envStr)
	if ok {
		return val
	}
	return ""

}
