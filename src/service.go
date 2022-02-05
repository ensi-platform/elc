package src

import (
	"errors"
	"fmt"
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

func (svc *Service) GetEnv() ([]string, error) {
	var env []string

	for key, value := range svc.Config.Variables {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	if svc.TplCfg != nil {
		env = append(env, svc.TplCfg.GetEnv()...)
		env = append(env, fmt.Sprintf("TPL_PATH=%s", svc.TplCfg.Path))
	}

	env = append(env, svc.SvcCfg.GetEnv()...)
	env = append(env, fmt.Sprintf("SVC_PATH=%s", svc.SvcCfg.Path))
	env = append(env, fmt.Sprintf("APP_NAME=%s", svc.SvcCfg.Name))
	env = append(env, fmt.Sprintf("COMPOSE_PROJECT_NAME=%s-%s", svc.Config.Name, svc.SvcCfg.Name))
	env = append(env, fmt.Sprintf("WORKSPACE_NAME=%s", svc.Config.Name))
	env = append(env, fmt.Sprintf("WORKSPACE_PATH=%s", svc.Config.WorkspacePath))

	return env, nil
}

func (svc *Service) GetComposeFile() (string, error) {
	if svc.SvcCfg.ComposeFile != "" {
		return svc.SvcCfg.ComposeFile, nil
	}

	if svc.TplCfg != nil && svc.TplCfg.ComposeFile != "" {
		return svc.TplCfg.ComposeFile, nil
	}

	return "", errors.New("compose file is not defined in service or template")
}

func (svc *Service) execComposeToString(composeCommand []string) (string, error) {
	env, err := svc.GetEnv()
	if err != nil {
		return "", err
	}

	composeFile, err := svc.GetComposeFile()
	if err != nil {
		return "", err
	}

	command := append([]string{"docker", "compose", "-f", composeFile}, composeCommand...)
	_, out, err := execToString(command, env)
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

	composeFile, err := svc.GetComposeFile()
	if err != nil {
		return 0, err
	}

	command := append([]string{"docker", "compose", "-f", composeFile}, composeCommand...)
	code, err := execInteractive(command, env)
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
	command = append(command, "app")
	command = append(command, params.Cmd...)
	code, err := svc.execComposeInteractive(command)
	if err != nil {
		return 0, err
	}

	return code, nil
}
