package core

import (
	"fmt"
	"os"
	"strings"
)

var hookNames = []string{
	"applypatch-msg",
	"pre-applypatch",
	"post-applypatch",
	"pre-commit",
	"pre-merge-commit",
	"prepare-commit-msg",
	"commit-msg",
	"post-commit",
	"pre-rebase",
	"post-checkout",
	"post-merge",
	"pre-push",
	"pre-receive",
	"update",
	"proc-receive",
	"post-receive",
	"post-update",
	"reference-transaction",
	"push-to-checkout",
	"pre-auto-gc",
	"post-rewrite",
	"sendemail-validate",
	"fsmonitor-watchman",
	"p4-changelist",
	"p4-prepare-changelist",
	"p4-post-changelist",
	"p4-pre-submit",
	"post-index-change",
}

var hookScript = `#!/bin/bash
set -e

ELC_BINARY="%s"
HOOKS_FOLDER="%s"
HOOK_NAME="%s"

if command -v $ELC_BINARY &> /dev/null; then
    $ELC_BINARY --mode=hook --no-tty $0
else
    for script in ./$HOOKS_FOLDER/$HOOK_NAME/* ; do
        if [ -f $script ]; then
            $script
        fi
    done
fi
`

func GenerateHookScripts(options *GlobalOptions, svcPath string, elcBinary string, scriptsFolder string) error {
	gitPath := fmt.Sprintf("%s/.git", svcPath)
	if Pc.FileExists(gitPath) == false {
		_, _ = Pc.Println(fmt.Sprintf("\033[0;33mRepository %s is not exists, skip hooks installation.\033[0m", gitPath))
		return nil
	}

	hooksPath := fmt.Sprintf("%s/hooks", gitPath)
	if Pc.FileExists(hooksPath) == false {
		if options.Debug {
			_, _ = Pc.Printf("mkdir %s\n", hooksPath)
		}

		if !options.DryRun {
			err := Pc.CreateDir(hooksPath)
			if err != nil {
				return err
			}
		}
	}

	scriptsFolder = strings.ReplaceAll(scriptsFolder, "./", "")
	scriptsFolder = strings.Trim(scriptsFolder, "/")

	scriptsFolderPath := fmt.Sprintf("%s/%s", svcPath, scriptsFolder)
	if Pc.FileExists(scriptsFolderPath) == false {
		_, _ = Pc.Println(fmt.Sprintf("\033[0;33mFolder %s is not exists, skip hooks installation.\033[0m", scriptsFolderPath))
		return nil
	}

	for _, hookName := range hookNames {
		scriptPath := fmt.Sprintf("%s/%s", hooksPath, hookName)
		scriptContent := fmt.Sprintf(hookScript, elcBinary, scriptsFolder, hookName)

		var filePermissions os.FileMode = 0775

		if Pc.FileExists(scriptPath) == false {
			if options.Debug {
				_, _ = Pc.Printf("touch %s\n", scriptPath)
				_, _ = Pc.Printf("chmod %s %s\n", filePermissions, scriptPath)
			}

			if !options.DryRun {
				err := Pc.CreateFile(scriptPath)
				if err != nil {
					return err
				}

				err = Pc.Chmod(scriptPath, filePermissions)
				if err != nil {
					return err
				}
			}
		}

		if options.Debug {
			_, _ = Pc.Printf("echo \"<script>\" > %s\n", scriptPath)
		}

		if !options.DryRun {
			err := Pc.WriteFile(scriptPath, []byte(scriptContent), filePermissions)
			if err != nil {
				return err
			}
		}
	}

	_, _ = Pc.Println(fmt.Sprintf("\033[0;32mFiles in %s updated.\033[0m", hooksPath))

	return nil
}
