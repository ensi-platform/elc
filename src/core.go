package src

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var Version string

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

func generateHookScript(scripts []string, elcBinary string) string {
	result := make([]string, 0)
	result = append(result, "#!/bin/bash")
	result = append(result, "set -e")
	result = append(result, `printf "\x1b[0;34m%s\x1b[39;49;00m\n" "Run hook in ELC"`)
	for _, script := range scripts {
		result = append(result, fmt.Sprintf("%s --mode=hook %s", elcBinary, script))
	}

	return strings.Join(result, "\n")
}
