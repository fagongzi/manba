package util

import (
	"fmt"
)

// set on build time
var (
	GitCommit = ""
	BuildTime = ""
	GoVersion = ""
	Version   = ""
)

// PrintVersion Print out version information
func PrintVersion() {
	fmt.Println("Version  : ", Version)
	fmt.Println("GitCommit: ", GitCommit)
	fmt.Println("BuildTime: ", BuildTime)
	fmt.Println("GoVersion: ", GoVersion)
}
