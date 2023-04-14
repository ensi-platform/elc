package core

import (
	"fmt"
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

func GenerateHookScripts(elcBinary string, hooksFolder string) error {
	for _, hookName := range hookNames {
		scriptContent := fmt.Sprintf(hookScript, elcBinary, hooksFolder, hookName)
		scriptPath := fmt.Sprintf(".git/hooks/%s", hookName)

		if Pc.FileExists(".git/hooks") == false {
			err := Pc.CreateDir(".git/hooks")
			if err != nil {
				return err
			}
		}

		if Pc.FileExists(scriptPath) == false {
			err := Pc.CreateFile(scriptPath)
			if err != nil {
				return err
			}

			err = Pc.Chmod(scriptPath, 0775)
			if err != nil {
				return err
			}
		}

		err := Pc.WriteFile(scriptPath, []byte(scriptContent), 0775)
		if err != nil {
			return err
		}
	}

	return nil
}
