package core

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var Version string

type GlobalOptions struct {
	WorkspaceName string
	ComponentName string
	Debug         bool
	Cmd           []string
	Force         bool
	Mode          string
	WorkingDir    string
	UID           int
	Tag           string
	DryRun        bool
	NoTty         bool
}

func contains(list []string, item string) bool {
	for _, value := range list {
		if value == item {
			return true
		}
	}
	return false
}

type reResult map[string]string

func reFindMaps(pattern string, subject string) ([]reResult, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	matches := re.FindAllStringSubmatch(subject, -1)
	names := re.SubexpNames()
	var result []reResult
	for _, match := range matches {
		foundFields := reResult{}
		for i, field := range match {
			if names[i] == "" {
				continue
			}
			foundFields[names[i]] = field
		}
		result = append(result, foundFields)
	}

	return result, nil
}

func substVars(expr string, ctx *Context) (string, error) {
	foundVars, err := reFindMaps(`\$\{(?P<name>[^:}]+)(:-(?P<value>[^}]+))?\}`, expr)
	if err != nil {
		return "", err
	}

	for _, foundVar := range foundVars {
		varName := foundVar["name"]
		value, found := ctx.find(varName)
		if !found {
			value, found = foundVar["value"]
			if !found {
				return "", errors.New(fmt.Sprintf("variable %s is not set", varName))
			}

			if strings.HasPrefix(value, "$") {
				varRef := strings.TrimLeft(value, "$")
				value, found = ctx.find(varRef)
				if !found {
					return "", errors.New(fmt.Sprintf("variable %s is not set", varRef))
				}
			}
		}
		re, err := regexp.Compile(fmt.Sprintf(`\$\{%s(?::-[^}]+)?\}`, varName))
		if err != nil {
			return "", err
		}
		expr = re.ReplaceAllString(expr, value)
	}

	return expr, nil
}
