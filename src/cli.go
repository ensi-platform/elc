package src

import (
	"fmt"
	"strings"
)

const CReset = "\033[0m"
const CYellow = "\033[33m"

func Color(text string, color string) string {
	return fmt.Sprintf("%s%s%s", color, text, CReset)
}

func NeedHelp(args []string, usage string, lines []string) bool {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
		fmt.Printf("Usage: %s %s\n", Pc.Args()[0], usage)
		if lines != nil {
			fmt.Println("")
			fmt.Println(strings.Join(lines, "\n"))
			fmt.Println("")
		}
		return true
	}
	return false
}
