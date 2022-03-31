package src

import (
	"errors"
	"fmt"
	"strconv"
)

type Component struct {
	Name        string
	Config      *ComponentConfig
	Template    *ComponentConfig
	JustStarted bool
	Context     *Context
	Workspace   *Workspace
}

func NewComponent(compName string, compCfg *ComponentConfig, ws *Workspace) *Component {
	return &Component{
		Name:      compName,
		Config:    compCfg,
		Workspace: ws,
	}
}

func (comp *Component) init() error {
	ctx := make(Context, 0)
	ctx.append(comp.Workspace.Context)

	ctx = ctx.add("APP_NAME", comp.Name)
	ctx = ctx.add("COMPOSE_PROJECT_NAME", fmt.Sprintf("%s-%s", comp.Workspace.Config.Name, comp.Name))
	svcPath, err := ctx.renderString(comp.Config.Path)
	if err != nil {
		return err
	}
	ctx = ctx.add("SVC_PATH", svcPath)

	if comp.Config.Extends != "" {
		tpl, found := comp.Workspace.Config.Components[comp.Config.Extends]
		if !found {
			return errors.New(fmt.Sprintf("template '%s' is not found", comp.Config.Extends))
		}
		comp.Template = &tpl

		tplPath, err := ctx.renderString(tpl.Path)
		if err != nil {
			return err
		}
		ctx = ctx.add("TPL_PATH", tplPath)
		if tpl.ComposeFile == "" {
			tpl.ComposeFile = "${TPL_PATH}/docker-compose.yml"
		}
		composeFile, err := ctx.renderString(tpl.ComposeFile)
		if err != nil {
			return err
		}
		ctx = ctx.add("COMPOSE_FILE", composeFile)
		for _, pair := range tpl.Variables {
			value, err := ctx.renderString(pair.Value.(string))
			if err != nil {
				return err
			}
			ctx = ctx.add(pair.Key.(string), value)
		}
	}

	if comp.Config.ComposeFile != "" {
		composeFile, err := ctx.renderString(comp.Config.ComposeFile)
		if err != nil {
			return err
		}
		ctx = ctx.add("COMPOSE_FILE", composeFile)
	}
	composeFile, found := ctx.find("COMPOSE_FILE")
	if !found || composeFile == "" {
		composeFile, err := ctx.renderString("${SVC_PATH}/docker-compose.yml")
		if err != nil {
			return err
		}
		ctx = ctx.add("COMPOSE_FILE", composeFile)
	}

	for _, pair := range comp.Config.Variables {
		value, err := ctx.renderString(pair.Value.(string))
		if err != nil {
			return err
		}
		ctx = ctx.add(pair.Key.(string), value)
	}

	comp.Context = &ctx

	return nil
}

func (comp *Component) execComposeToString(composeCommand []string) (string, error) {
	composeFile, _ := comp.Context.find("COMPOSE_FILE")
	command := append([]string{"docker", "compose", "-f", composeFile}, composeCommand...)
	_, out, err := Pc.ExecToString(command, comp.Context.renderMapToEnv())
	if err != nil {
		return "", err
	}

	return out, nil
}

func (comp *Component) execComposeInteractive(composeCommand []string) (int, error) {
	composeFile, _ := comp.Context.find("COMPOSE_FILE")
	command := append([]string{"docker", "compose", "-f", composeFile}, composeCommand...)
	code, err := Pc.ExecInteractive(command, comp.Context.renderMapToEnv())
	if err != nil {
		return 0, err
	}

	return code, nil
}

func (comp *Component) IsRunning() (bool, error) {
	out, err := comp.execComposeToString([]string{"ps", "--status=running", "-q"})
	if err != nil {
		return false, err
	}

	return out != "", nil
}

type SvcStartParams struct {
	Force bool
	Mode  string
}

func (comp *Component) Start(params *SvcStartParams) error {
	if comp.JustStarted {
		return nil
	}

	running, err := comp.IsRunning()
	if err != nil {
		return err
	}

	if !running || params.Force {
		err := comp.startDependencies(params)
		if err != nil {
			return err
		}
	}

	if !running {
		_, err = comp.execComposeInteractive([]string{"up", "-d"})
		if err != nil {
			return err
		}
	}

	return nil
}

func (comp *Component) startDependencies(params *SvcStartParams) error {
	for _, depName := range comp.Config.GetDeps(params.Mode) {
		depComp, found := comp.Workspace.Components[depName]
		if !found {
			return errors.New(fmt.Sprintf("dependency with name '%s' is not defined", depName))
		}
		err := depComp.Start(params)
		if err != nil {
			return err
		}
	}

	return nil
}

func (comp *Component) Stop() error {
	running, err := comp.IsRunning()
	if err != nil {
		return err
	}
	if running {
		_, err = comp.execComposeInteractive([]string{"stop"})
		if err != nil {
			return err
		}
	}

	return nil
}

func (comp *Component) Destroy() error {
	running, err := comp.IsRunning()
	if err != nil {
		return err
	}
	if running {
		_, err := comp.execComposeInteractive([]string{"down"})
		if err != nil {
			return err
		}
	}

	return nil
}

type SvcRestartParams struct {
	Hard bool
}

func (comp *Component) Restart(params *SvcRestartParams) error {
	var err error
	if params.Hard {
		err = comp.Destroy()
		if err != nil {
			return err
		}
	} else {
		err = comp.Stop()
		if err != nil {
			return err
		}
	}
	err = comp.Start(&SvcStartParams{})
	if err != nil {
		return err
	}

	return nil
}

type SvcComposeParams struct {
	Cmd     []string
	SvcName string
}

func (comp *Component) Compose(params *SvcComposeParams) (int, error) {
	code, err := comp.execComposeInteractive(params.Cmd)
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

func (comp *Component) Exec(params *SvcExecParams) (int, error) {
	err := comp.Start(&params.SvcStartParams)
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

	if !Pc.IsTerminal() {
		command = append(command, "-T")
	}
	command = append(command, "app")

	command = append(command, params.Cmd...)
	code, err := comp.execComposeInteractive(command)
	if err != nil {
		return 0, err
	}

	return code, nil
}

func (comp *Component) DumpVars() error {
	for _, line := range comp.Context.renderMapToEnv() {
		_, _ = Pc.Println(line)
	}

	return nil
}
