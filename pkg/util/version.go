package util

import (
	"fmt"
)

// set on build time
var (
	GitCommit = ""
	BuildTime = ""
	GoVersion = ""
)

// PrintVersion 输出版本信息
func PrintVersion() bool {
	fmt.Println("GitCommit: ", GitCommit)
	fmt.Println("BuildTime: ", BuildTime)
	fmt.Println("GoVersion: ", GoVersion)

	return true
}
