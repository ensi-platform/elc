package src

import "fmt"

func UpdateBinaryAction(version string) error {
	env := make([]string, 0)
	if version != "" {
		env = append(env, fmt.Sprintf("VERSION=%s", version))
	}

	hc, err := checkAndLoadHC()
	if err != nil {
		return err
	}

	_, err = Pc.ExecInteractive([]string{"bash", "-c", hc.UpdateCommand}, env)
	if err != nil {
		return err
	}

	return nil
}

func FixUpdateBinaryCommandAction() error {
	hc, err := checkAndLoadHC()
	if err != nil {
		return err
	}

	hc.UpdateCommand = defaultUpdateCommand
	err = SaveHomeConfig(hc)
	if err != nil {
		return err
	}

	return nil
}
