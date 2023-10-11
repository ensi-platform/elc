package actions

import (
	"fmt"
	"github.com/ensi-platform/elc/core"
)

func UpdateBinaryAction(version string) error {
	env := make([]string, 0)
	if version != "" {
		env = append(env, fmt.Sprintf("VERSION=%s", version))
	}

	hc, err := core.CheckAndLoadHC()
	if err != nil {
		return err
	}

	_, err = core.Pc.ExecInteractive([]string{"bash", "-c", hc.UpdateCommand}, env)
	if err != nil {
		return err
	}

	return nil
}

func FixUpdateBinaryCommandAction() error {
	hc, err := core.CheckAndLoadHC()
	if err != nil {
		return err
	}

	hc.UpdateCommand = core.DefaultUpdateCommand
	err = core.SaveHomeConfig(hc)
	if err != nil {
		return err
	}

	return nil
}
