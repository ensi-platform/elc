package main

import (
	"github.com/madridianfox/elc/cmd"
	"os"
)

func main() {
	command := cmd.InitCobra()
	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}
