package src

import (
	"errors"
	"fmt"
	"regexp"
)

func contains(list []string, item string) bool {
	for _, value := range list {
		if value == item {
			return true
		}
	}
	return false
}

func renderMapToEnv(env map[string]string) []string {
	var result []string
	for key, value := range env {
		result = append(result, fmt.Sprintf("%s=%s", key, value))
	}

	return result
}

type reResult map[string]string

func reFindMaps(pattern string, subject string) ([]reResult, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	matches := re.FindAllStringSubmatch(subject, -1)
	names := re.SubexpNames()
	result := []reResult{}
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

func substVars(expr string, vars map[string]string) (string, error) {
	foundVars, err := reFindMaps(`\$\{(?P<name>[^:}]+)(:-(?P<value>[^}]+))?\}`, expr)
	if err != nil {
		return "", err
	}

	for _, foundVar := range foundVars {
		varName := foundVar["name"]
		value, found := vars[varName]
		if !found {
			value, found = foundVar["value"]
			if !found {
				return "", errors.New(fmt.Sprintf("variable %s is not set", varName))
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
