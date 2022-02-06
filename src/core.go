package src

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
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

func SetGitHooks(scriptsFolder string, elcBinary string) error {
	folders, err := ioutil.ReadDir(scriptsFolder)
	if err != nil {
		return err
	}
	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}
		files, err := ioutil.ReadDir(path.Join(scriptsFolder, folder.Name()))
		if err != nil {
			return err
		}
		hookScripts := make([]string, 0)
		for _, file := range files {
			hookScripts = append(hookScripts, path.Join(scriptsFolder, folder.Name(), file.Name()))
		}
		script := generateHookScript(hookScripts, elcBinary)
		err = ioutil.WriteFile(fmt.Sprintf(".git/hooks/%s", folder.Name()), []byte(script), 0755)
		if err != nil {
			return err
		}
	}

	return nil
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
