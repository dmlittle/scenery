package main

import "github.com/dmlittle/scenery/pkg/cmd"

// Version is updated by linker flags during build time
var Version = ""

func main() {
	cmd.Execute(Version)
}
