package main

import (
	elc "github.com/madridianfox/elc/src"
	"os"
)

func main() {
	command := elc.InitCobra()
	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}
