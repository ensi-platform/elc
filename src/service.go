package src

import (
	"errors"
	"fmt"
	"github.com/mattn/go-isatty"
	"os"
	"strconv"
)

type Service struct {
	Config *MainConfig
	SvcCfg *ServiceConfig
	TplCfg *TemplateConfig
}

func CreateFromSvcName(cfg *MainConfig, svcName string) (*Service, error) {
	sts := Service{Config: cfg}
	svc, err := cfg.FindServiceByName(svcName)
	if err != nil {
		return nil, err
	}
	sts.SvcCfg = svc

	if svc.Extends != "" {
		tpl, err := cfg.FindTemplateByName(svc.Extends)
		if err != nil {
			return nil, err
		}
		sts.TplCfg = tpl
	}
	return &sts, nil
}

func (svc *Service) GetEnv() (map[string]string, error) {
	env, err := svc.Config.makeGlobalEnv()
	if err != nil {
		return nil, err
	}

	env["APP_NAME"] = svc.SvcCfg.Name
	env["COMPOSE_PROJECT_NAME"] = fmt.Sprintf("%s-%s", svc.Config.Name, svc.SvcCfg.Name)

	env["SVC_PATH"], err = substVars(svc.SvcCfg.Path, env)
	if err != nil {
		return nil, err
	}

	if svc.TplCfg != nil {
		env["TPL_PATH"], err = substVars(svc.TplCfg.Path, env)
		if err != nil {
			return nil, err
		}
		env["COMPOSE_FILE"], err = substVars(svc.TplCfg.ComposeFile, env)
		if err != nil {
			return nil, err
		}
		for key, value := range svc.TplCfg.Variables {
			env[key], err = substVars(value, env)
			if err != nil {
				return nil, err
			}
		}
	}

	if svc.SvcCfg.ComposeFile != "" {
		env["COMPOSE_FILE"], err = substVars(svc.SvcCfg.ComposeFile, env)
		if err != nil {
			return nil, err
		}
	}
	for key, value := range svc.SvcCfg.Variables {
		env[key], err = substVars(value, env)
		if err != nil {
			return nil, err
		}
	}

	return env, nil
}

func (svc *Service) execComposeToString(composeCommand []string) (string, error) {
	env, err := svc.GetEnv()
	if err != nil {
		return "", err
	}

	composeFile, found := env["COMPOSE_FILE"]
	if !found {
		return "", errors.New("compose file is not defined in service or template")
	}

	command := append([]string{"docker", "compose", "-f", composeFile}, composeCommand...)
	_, out, err := execToString(command, renderMapToEnv(env))
	if err != nil {
		return "", err
	}

	return out, nil
}

func (svc *Service) execComposeInteractive(composeCommand []string) (int, error) {
	env, err := svc.GetEnv()
	if err != nil {
		return 0, err
	}

	composeFile, found := env["COMPOSE_FILE"]
	if !found {
		return 0, errors.New("compose file is not defined in service or template")
	}

	command := append([]string{"docker", "compose", "-f", composeFile}, composeCommand...)
	code, err := execInteractive(command, renderMapToEnv(env))
	if err != nil {
		return 0, err
	}

	return code, nil
}

func (svc *Service) IsRunning() (bool, error) {
	out, err := svc.execComposeToString([]string{"ps", "--status=running", "-q"})
	if err != nil {
		return false, err
	}

	return out != "", nil
}

type SvcStartParams struct {
	Force bool
	Tag   string
}

func (svc *Service) Start(params *SvcStartParams) error {
	svc.Config.WillStart = append(svc.Config.WillStart, svc.SvcCfg.Name)

	running, err := svc.IsRunning()
	if err != nil {
		return err
	}

	if !running || params.Force {
		err := svc.startDependencies(params)
		if err != nil {
			return err
		}
	}

	if !running {
		_, err = svc.execComposeInteractive([]string{"up", "-d"})
		if err != nil {
			return err
		}
	}

	return nil
}

func (svc *Service) startDependencies(params *SvcStartParams) error {
	for _, depName := range svc.SvcCfg.GetDeps(params.Tag) {
		if contains(svc.Config.WillStart, depName) {
			continue
		}

		depSvc, err := CreateFromSvcName(svc.Config, depName)
		if err != nil {
			return err
		}

		err = depSvc.Start(params)
		if err != nil {
			return err
		}
	}

	return nil
}

func (svc *Service) Stop() error {
	_, err := svc.execComposeInteractive([]string{"stop"})
	if err != nil {
		return err
	}

	return nil
}

func (svc *Service) Destroy() error {
	_, err := svc.execComposeInteractive([]string{"down"})
	if err != nil {
		return err
	}

	return nil
}

type SvcRestartParams struct {
	Hard bool
}

func (svc *Service) Restart(params *SvcRestartParams) error {
	var err error
	if params.Hard {
		err = svc.Destroy()
		if err != nil {
			return err
		}
	} else {
		err = svc.Stop()
		if err != nil {
			return err
		}
	}
	err = svc.Start(&SvcStartParams{})
	if err != nil {
		return err
	}

	return nil
}

type SvcComposeParams struct {
	Cmd     []string
	SvcName string
}

func (svc *Service) Compose(params *SvcComposeParams) (int, error) {
	code, err := svc.execComposeInteractive(params.Cmd)
	if err != nil {
		return 0, err
	}

	return code, nil
}

type SvcExecParams struct {
	SvcComposeParams
	SvcStartParams
	WorkingDir string
	UID        int
}

func (svc *Service) Exec(params *SvcExecParams) (int, error) {
	err := svc.Start(&params.SvcStartParams)
	if err != nil {
		return 0, err
	}

	command := []string{"exec"}
	if params.WorkingDir != "" {
		command = append(command, "-w", params.WorkingDir)
	}
	if params.UID > -1 {
		command = append(command, "-u", strconv.Itoa(params.UID))
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		command = append(command, "-T")
	}
	command = append(command, "app")

	command = append(command, params.Cmd...)
	code, err := svc.execComposeInteractive(command)
	if err != nil {
		return 0, err
	}

	return code, nil
}

func (svc *Service) DumpVars() error {
	env, err := svc.GetEnv()
	if err != nil {
		return err
	}

	for _, line := range renderMapToEnv(env) {
		fmt.Println(line)
	}

	return nil
}
