package utils

import (
	"encoding/json"
	"runtime"
)

const (
	OSWindows = "windows"
	OSLinux   = "linux"
	OSMac     = "darwin"
)

func GetOS() string {
	return runtime.GOOS
}

func IsWindows() bool {
	return GetOS() == OSWindows
}

func ToDebugString(v interface{}) string {
	str, _ := json.Marshal(v)
	return string(str)
}
