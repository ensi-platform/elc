package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	ctx := make(Context, len(*comp.Workspace.Context))
	copy(ctx, *comp.Workspace.Context)

	ctx = ctx.add("APP_NAME", comp.Name)
	ctx = ctx.add("COMPOSE_PROJECT_NAME", fmt.Sprintf("%s-%s", comp.Workspace.Config.Name, comp.Name))
	svcPath, err := ctx.RenderString(comp.Config.Path)
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

		tplPath, err := ctx.RenderString(tpl.Path)
		if err != nil {
			return err
		}
		ctx = ctx.add("TPL_PATH", tplPath)
		if tpl.ComposeFile == "" {
			tpl.ComposeFile = "${TPL_PATH}/docker-compose.yml"
		}
		composeFile, err := ctx.RenderString(tpl.ComposeFile)
		if err != nil {
			return err
		}
		ctx = ctx.add("COMPOSE_FILE", composeFile)
		for _, pair := range tpl.Variables {
			value, err := ctx.RenderString(pair.Value.(string))
			if err != nil {
				return err
			}
			ctx = ctx.add(pair.Key.(string), value)
		}
	}

	if comp.Config.ComposeFile != "" {
		composeFile, err := ctx.RenderString(comp.Config.ComposeFile)
		if err != nil {
			return err
		}
		ctx = ctx.add("COMPOSE_FILE", composeFile)
	}
	composeFile, found := ctx.find("COMPOSE_FILE")
	if !found || composeFile == "" {
		composeFile, err := ctx.RenderString("${SVC_PATH}/docker-compose.yml")
		if err != nil {
			return err
		}
		ctx = ctx.add("COMPOSE_FILE", composeFile)
	}

	for _, pair := range comp.Config.Variables {
		value, err := ctx.RenderString(pair.Value.(string))
		if err != nil {
			return err
		}
		ctx = ctx.add(pair.Key.(string), value)
	}

	comp.Context = &ctx

	return nil
}

func (comp *Component) execComposeToString(composeCommand []string, options *GlobalOptions) (string, error) {
	composeFile, _ := comp.Context.find("COMPOSE_FILE")
	command := append([]string{"docker", "compose", "-f", composeFile}, composeCommand...)

	if options.Debug {
		_, _ = Pc.Printf(">> %s\n", strings.Join(command, " "))
	}

	if !options.DryRun {
		_, out, err := Pc.ExecToString(command, comp.Context.renderMapToEnv())
		if err != nil {
			return "", err
		}
		return out, nil
	}

	return "", nil
}

func (comp *Component) execComposeInteractive(composeCommand []string, options *GlobalOptions) (int, error) {
	composeFile, _ := comp.Context.find("COMPOSE_FILE")
	command := append([]string{"docker", "compose", "-f", composeFile}, composeCommand...)

	if options.Debug {
		_, _ = Pc.Printf(">> %s\n", strings.Join(command, " "))
	}

	if !options.DryRun {
		code, err := Pc.ExecInteractive(command, comp.Context.renderMapToEnv())
		if err != nil {
			return 0, err
		}

		return code, nil
	}

	return 0, nil
}

func (comp *Component) execInteractive(command []string, options *GlobalOptions) (int, error) {
	if options.Debug {
		_, _ = Pc.Printf(">> %s\n", strings.Join(command, " "))
	}

	if !options.DryRun {
		code, err := Pc.ExecInteractive(command, comp.Context.renderMapToEnv())
		if err != nil {
			return 0, err
		}
		return code, nil
	}

	return 0, nil
}

func (comp *Component) IsRunning(options *GlobalOptions) (bool, error) {
	out, err := comp.execComposeToString([]string{"ps", "--status=running", "-q"}, options)
	if err != nil {
		return false, err
	}

	return out != "", nil
}

func (comp *Component) IsCloned() (bool, error) {
	svcPath, found := comp.Context.find("SVC_PATH")
	if !found {
		return false, errors.New("path of component is not defined.Check workspace.yaml")
	}

	return Pc.FileExists(svcPath), nil
}

func (comp *Component) Start(options *GlobalOptions) error {
	if comp.JustStarted {
		return nil
	}

	cloned, err := comp.IsCloned()
	if err != nil {
		return err
	}

	if !cloned {
		_, _ = Pc.Println("component is not cloned")
		return nil
	}

	running, err := comp.IsRunning(options)
	if err != nil {
		return err
	}

	if !running || options.Force {
		err := comp.startDependencies(options)
		if err != nil {
			return err
		}
	}

	if !running {
		_, err = comp.execComposeInteractive([]string{"up", "-d"}, options)
		if err != nil {
			return err
		}
	}

	return nil
}

func (comp *Component) startDependencies(params *GlobalOptions) error {
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

func (comp *Component) Stop(options *GlobalOptions) error {
	cloned, err := comp.IsCloned()
	if err != nil {
		return err
	}

	if !cloned {
		_, _ = Pc.Println("component is not cloned")
		return nil
	}

	running, err := comp.IsRunning(options)
	if err != nil {
		return err
	}
	if running {
		_, err = comp.execComposeInteractive([]string{"stop"}, options)
		if err != nil {
			return err
		}
	}

	return nil
}

func (comp *Component) Destroy(options *GlobalOptions) error {
	cloned, err := comp.IsCloned()
	if err != nil {
		return err
	}

	if !cloned {
		_, _ = Pc.Println("component is not cloned")
		return nil
	}

	running, err := comp.IsRunning(options)
	if err != nil {
		return err
	}
	if running {
		_, err := comp.execComposeInteractive([]string{"down"}, options)
		if err != nil {
			return err
		}
	}

	return nil
}

func (comp *Component) Restart(hard bool, options *GlobalOptions) error {
	var err error

	if hard {
		err = comp.Destroy(options)
		if err != nil {
			return err
		}
	} else {
		err = comp.Stop(options)
		if err != nil {
			return err
		}
	}
	err = comp.Start(&GlobalOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (comp *Component) Compose(params *GlobalOptions) (int, error) {
	cloned, err := comp.IsCloned()
	if err != nil {
		return 1, err
	}

	if !cloned {
		_, _ = Pc.Println("component is not cloned")
		return 1, nil
	}

	code, err := comp.execComposeInteractive(params.Cmd, params)
	if err != nil {
		return 0, err
	}

	return code, nil
}

func (comp *Component) Exec(options *GlobalOptions) (int, error) {
	err := comp.Start(options)
	if err != nil {
		return 0, err
	}

	command := []string{"exec"}
	if options.WorkingDir != "" {
		command = append(command, "-w", options.WorkingDir)
	}
	if options.UID > -1 {
		command = append(command, "-u", strconv.Itoa(options.UID))
	} else {
		userId, found := comp.Context.find("USER_ID")
		if !found {
			return 0, errors.New("variable USER_ID is not defined")
		}

		groupId, found := comp.Context.find("GROUP_ID")
		if !found {
			return 0, errors.New("variable USER_ID is not defined")
		}

		command = append(command, "-u", fmt.Sprintf("%s:%s", userId, groupId))
	}

	if options.NoTty || !Pc.IsTerminal() {
		command = append(command, "-T")
	}
	command = append(command, "app")

	command = append(command, options.Cmd...)
	code, err := comp.execComposeInteractive(command, options)
	if err != nil {
		return 0, err
	}

	return code, nil
}

func (comp *Component) Run(options *GlobalOptions) (int, error) {
	cloned, err := comp.IsCloned()
	if err != nil {
		return 1, err
	}

	if !cloned {
		_, _ = Pc.Println("component is not cloned")
		return 1, nil
	}

	command := []string{"run", "--rm", "--entrypoint=''"}
	if options.WorkingDir != "" {
		command = append(command, "-w", options.WorkingDir)
	}
	if options.UID > -1 {
		command = append(command, "-u", strconv.Itoa(options.UID))
	} else {
		userId, found := comp.Context.find("USER_ID")
		if !found {
			return 0, errors.New("variable USER_ID is not defined")
		}

		groupId, found := comp.Context.find("GROUP_ID")
		if !found {
			return 0, errors.New("variable USER_ID is not defined")
		}

		command = append(command, "-u", fmt.Sprintf("%s:%s", userId, groupId))
	}

	if options.NoTty || !Pc.IsTerminal() {
		command = append(command, "-T")
	}
	command = append(command, "app")

	command = append(command, options.Cmd...)
	code, err := comp.execComposeInteractive(command, options)
	if err != nil {
		return 0, err
	}

	return code, nil
}

func (comp *Component) Wrap(command []string, options *GlobalOptions) (int, error) {
	code, err := comp.execInteractive(command, options)
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

func (comp *Component) getAfterCloneHook() string {
	if comp.Config.AfterCloneHook != "" {
		return comp.Config.AfterCloneHook
	}

	if comp.Template != nil {
		return comp.Template.AfterCloneHook
	}

	return ""
}

func (comp *Component) Clone(options *GlobalOptions, noHook bool) error {
	cloned, err := comp.IsCloned()
	if err != nil {
		return err
	}

	if cloned {
		_, _ = Pc.Println("component is already cloned")
		return nil
	}

	if comp.Config.Repository == "" {
		return errors.New(fmt.Sprintf("repository of component %s is not defined. Check workspace.yaml", comp.Name))
	}
	svcPath, found := comp.Context.find("SVC_PATH")
	if !found {
		return errors.New("path of component is not defined.Check workspace.yaml")
	}
	if Pc.FileExists(svcPath) {
		_, _ = Pc.Printf("Folder of component %s already exists. Skip.\n", comp.Name)
		return nil
	} else {
		_, err := comp.execInteractive([]string{"git", "clone", comp.Config.Repository, svcPath}, options)
		if err != nil {
			return err
		}

		if !noHook {
			afterCloneHook := comp.getAfterCloneHook()
			if afterCloneHook == "" {
				return nil
			}

			afterCloneHook, err = comp.Context.RenderString(afterCloneHook)
			if err != nil {
				return err
			}

			if afterCloneHook != "" {
				_, err = comp.execInteractive([]string{afterCloneHook}, options)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func (comp *Component) UpdateHooks(options *GlobalOptions, elcBinary string, scriptsFolder string) error {
	svcPath, found := comp.Context.find("SVC_PATH")
	if !found {
		return errors.New("path of component is not defined.Check workspace.yaml")
	}

	return GenerateHookScripts(options, svcPath, elcBinary, scriptsFolder)
}
