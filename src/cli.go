package src

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

type DefaultOptions struct {
	ComponentName string
}

var defaultOptions DefaultOptions
var startParams SvcStartParams

func parseStartParams(cmd *cobra.Command) {
	startParams = SvcStartParams{}
	cmd.Flags().BoolVar(&startParams.Force, "force", false, "force start dependencies, even if service already started")
	cmd.Flags().StringVar(&startParams.Mode, "mode", "default", "start only dependencies with specified mode, by default starts 'default' dependencies")
}

func InitCobra() *cobra.Command {
	defaultOptions = DefaultOptions{}
	var rootCmd = &cobra.Command{
		Use:  "elc",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Hello, cobra!")
			fmt.Printf("> %+v\n", args)
			fmt.Printf("> %+v\n", defaultOptions)
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&defaultOptions.ComponentName, "svc", "", "name of current component (deprecated)")
	rootCmd.PersistentFlags().StringVar(&defaultOptions.ComponentName, "component", "", "name of current component")
	rootCmd.Flags().SetInterspersed(false)

	NewWorkspaceCommand(rootCmd)
	NewServiceStartCommand(rootCmd)
	NewServiceStopCommand(rootCmd)
	NewServiceDestroyCommand(rootCmd)
	NewServiceRestartCommand(rootCmd)
	NewServiceVarsCommand(rootCmd)
	NewServiceComposeCommand(rootCmd)
	NewServiceWrapCommand(rootCmd)
	NewServiceExecCommand(rootCmd)
	NewServiceSetHooksCommand(rootCmd)
	NewUpdateCommand(rootCmd)
	NewFixUpdateCommand(rootCmd)

	return rootCmd
}

func NewWorkspaceCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use: "workspace",
	}
	NewWorkspaceListCommand(command)
	NewWorkspaceAddCommand(command)
	NewWorkspaceShowCommand(command)
	NewWorkspaceSelectCommand(command)
	parentCommand.AddCommand(command)
}

func NewWorkspaceListCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "list",
		Short: "Show list of registered workspaces",
		Long:  "Show list of registered workspaces.",
		RunE: func(cmd *cobra.Command, args []string) error {
			Pc = &RealPC{}

			hc, err := checkAndLoadHC()
			if err != nil {
				return err
			}
			for _, workspace := range hc.Workspaces {
				_, _ = Pc.Printf("%-10s %s\n", workspace.Name, workspace.Path)
			}
			return nil
		},
	}
	parentCommand.AddCommand(command)
}

func NewWorkspaceAddCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "add",
		Short: "Register new workspace",
		Long:  "Register new workspace.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			Pc = &RealPC{}

			hc, err := checkAndLoadHC()
			if err != nil {
				return err
			}

			name := args[0]
			wsPath := args[1]

			ws := hc.findWorkspace(name)
			if ws != nil {
				return errors.New(fmt.Sprintf("workspace with name '%s' already exists", name))
			}

			err = hc.AddWorkspace(name, wsPath)
			if err != nil {
				return err
			}

			_, _ = Pc.Printf("workspace '%s' is added\n", name)

			if hc.CurrentWorkspace == "" {
				hc.CurrentWorkspace = name
				err = SaveHomeConfig(hc)
				if err != nil {
					return err
				}

				_, _ = Pc.Printf("active workspace changed to '%s'\n", name)
			}

			return nil
		},
	}
	parentCommand.AddCommand(command)
}

func NewWorkspaceShowCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "show",
		Short: "Print current workspace name",
		Long:  "Print current workspace name.",
		RunE: func(cmd *cobra.Command, args []string) error {
			Pc = &RealPC{}

			hc, err := checkAndLoadHC()
			if err != nil {
				return err
			}
			_, _ = Pc.Println(hc.CurrentWorkspace)

			return nil
		},
	}
	parentCommand.AddCommand(command)
}

func NewWorkspaceSelectCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "select",
		Short: "Set current workspace",
		Long:  "Set workspace with name NAME as current.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			Pc = &RealPC{}

			hc, err := checkAndLoadHC()
			if err != nil {
				return err
			}
			name := args[0]

			ws := hc.findWorkspace(name)
			if ws == nil {
				return errors.New(fmt.Sprintf("workspace with name '%s' is not defined", name))
			}

			hc.CurrentWorkspace = name
			err = SaveHomeConfig(hc)
			if err != nil {
				return err
			}

			_, _ = Pc.Printf("active workspace changed to '%s'\n", name)

			return nil
		},
	}
	parentCommand.AddCommand(command)
}

func NewServiceStartCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "start",
		Short: "Start one or more services",
		Long:  "By default starts service found with current directory, but you can pass one or more service names instead.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			Pc = &RealPC{}

			ws, err := getWorkspaceConfig()
			if err != nil {
				return err
			}

			svcNames := args
			if len(svcNames) > 0 {
				for _, svcName := range svcNames {
					comp, err := ws.componentByName(svcName)
					if err != nil {
						return err
					}

					err = comp.Start(&startParams)
					if err != nil {
						return err
					}
				}
			} else {
				comp, err := ws.componentByPath()
				if err != nil {
					return err
				}

				err = comp.Start(&startParams)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
	parseStartParams(command)
	parentCommand.AddCommand(command)
}

func NewServiceStopCommand(parentCommand *cobra.Command) {
	var stopAll bool
	var command = &cobra.Command{
		Use:   "stop",
		Short: "Stop one or more services",
		Long:  "By default stops service found with current directory, but you can pass one or more service names instead.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			Pc = &RealPC{}

			ws, err := getWorkspaceConfig()
			if err != nil {
				return err
			}

			var svcNames []string
			if stopAll {
				svcNames = ws.getComponentNames()
			} else {
				svcNames = args
			}

			if len(svcNames) > 0 {
				for _, svcName := range svcNames {
					comp, err := ws.componentByName(svcName)
					if err != nil {
						return err
					}
					err = comp.Stop()
					if err != nil {
						return err
					}
				}
			} else {
				comp, err := ws.componentByPath()
				if err != nil {
					return err
				}

				err = comp.Stop()
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
	command.Flags().BoolVar(&stopAll, "all", false, "stop all services")
	parentCommand.AddCommand(command)
}

func NewServiceDestroyCommand(parentCommand *cobra.Command) {
	var destroyAll bool
	var command = &cobra.Command{
		Use:   "destroy",
		Short: "Stop and remove containers of one or more services",
		Long:  "By default destroys service found with current directory, but you can pass one or more service names instead.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			Pc = &RealPC{}

			ws, err := getWorkspaceConfig()
			if err != nil {
				return err
			}

			var svcNames []string
			if destroyAll {
				svcNames = ws.getComponentNames()
			} else {
				svcNames = args
			}

			if len(svcNames) > 0 {
				for _, svcName := range svcNames {
					comp, err := ws.componentByName(svcName)
					if err != nil {
						return err
					}

					err = comp.Destroy()
					if err != nil {
						return err
					}
				}
			} else {
				comp, err := ws.componentByPath()
				if err != nil {
					return err
				}

				err = comp.Destroy()
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
	command.Flags().BoolVar(&destroyAll, "all", false, "destroy all services")
	parentCommand.AddCommand(command)
}

func NewServiceRestartCommand(parentCommand *cobra.Command) {
	var hardRestart bool
	var command = &cobra.Command{
		Use:   "restart",
		Short: "Restart one or more services",
		Long:  "By default restart service found with current directory, but you can pass one or more service names instead.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			Pc = &RealPC{}

			ws, err := getWorkspaceConfig()
			if err != nil {
				return err
			}

			svcNames := args
			if len(svcNames) > 0 {
				for _, svcName := range svcNames {
					comp, err := ws.componentByName(svcName)
					if err != nil {
						return err
					}

					err = comp.Restart(hardRestart)
					if err != nil {
						return err
					}
				}
			} else {
				comp, err := ws.componentByPath()
				if err != nil {
					return err
				}

				err = comp.Restart(hardRestart)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
	command.Flags().BoolVar(&hardRestart, "all", false, "destroy container instead of stop it before start")
	parentCommand.AddCommand(command)
}

func NewServiceVarsCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "vars",
		Short: "Print all variables computed for service",
		Long:  "By default uses service found with current directory, but you can pass name of another service instead.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			Pc = &RealPC{}

			ws, err := getWorkspaceConfig()
			if err != nil {
				return err
			}

			var comp *Component

			if len(args) > 0 {
				comp, err = ws.componentByName(args[0])
				if err != nil {
					return err
				}
			} else {
				comp, err = ws.componentByPath()
				if err != nil {
					return err
				}
			}

			err = comp.DumpVars()
			if err != nil {
				return err
			}

			return nil
		},
	}
	parentCommand.AddCommand(command)
}

func NewServiceComposeCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "compose",
		Short: "Run docker-compose command",
		Long:  "By default uses service found with current directory.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			Pc = &RealPC{}

			ws, err := getWorkspaceConfig()
			if err != nil {
				return err
			}

			composeParams := SvcComposeParams{
				Cmd:     args,
				SvcName: defaultOptions.ComponentName,
			}

			if composeParams.SvcName == "" {
				composeParams.SvcName, err = ws.componentNameByPath()
				if err != nil {
					return err
				}
			}

			comp, err := ws.componentByName(composeParams.SvcName)
			if err != nil {
				return err
			}

			_, err = comp.Compose(&composeParams)
			if err != nil {
				return err
			}

			return nil
		},
	}
	command.Flags().SetInterspersed(false)
	parentCommand.AddCommand(command)
}

func NewServiceWrapCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "wrap",
		Short: "Execute command on host with env variables for service. For module uses variables of linked service",
		Long:  "By default uses service/module found with current directory.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			Pc = &RealPC{}

			ws, err := getWorkspaceConfig()
			if err != nil {
				return err
			}

			var comp *Component

			svcName := defaultOptions.ComponentName

			if svcName == "" {
				comp, err = ws.componentByPath()
			} else {
				comp, err = ws.componentByName(svcName)
			}
			if err != nil {
				return err
			}

			if comp.Config.HostedIn != "" {
				svcName = comp.Config.HostedIn
			} else {
				svcName = comp.Name
			}

			hostComp, err := ws.componentByName(svcName)
			if err != nil {
				return err
			}

			_, err = hostComp.Wrap(args)
			if err != nil {
				return err
			}

			return nil
		},
	}
	parentCommand.AddCommand(command)
}

func NewServiceExecCommand(parentCommand *cobra.Command) {
	var uid int
	var command = &cobra.Command{
		Use:   "exec",
		Short: "Execute command in container. For module uses container of linked service",
		Long:  "By default uses service/module found with current directory. Starts service if it is not running.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			Pc = &RealPC{}

			ws, err := getWorkspaceConfig()
			if err != nil {
				return err
			}

			execParams := SvcExecParams{}
			execParams.Cmd = args
			execParams.SvcName = defaultOptions.ComponentName
			execParams.Force = startParams.Force
			execParams.Mode = startParams.Mode
			execParams.UID = uid

			var comp *Component

			if execParams.SvcName == "" {
				comp, err = ws.componentByPath()
			} else {
				comp, err = ws.componentByName(execParams.SvcName)
			}
			if err != nil {
				return err
			}

			if comp.Config.HostedIn != "" {
				execParams.SvcName = comp.Config.HostedIn
			} else {
				execParams.SvcName = comp.Name
			}

			if comp.Config.ExecPath != "" {
				execParams.WorkingDir, err = ws.Context.renderString(comp.Config.ExecPath)
				if err != nil {
					return err
				}
			}

			hostComp, err := ws.componentByName(execParams.SvcName)
			if err != nil {
				return err
			}

			_, err = hostComp.Exec(&execParams)
			if err != nil {
				return err
			}

			return nil
		},
	}
	command.Flags().SetInterspersed(false)
	parseStartParams(command)
	command.Flags().IntVar(&uid, "uid", -1, "use another uid, by default uses uid of current user")
	parentCommand.AddCommand(command)
}

func NewServiceSetHooksCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "set-hooks",
		Short: "Install hooks from specified folder to .git/hooks",
		Long:  "HOOKS_PATH must contain subdirectories with names as git hooks, eg. 'pre-commit'.\nOne subdirectory can contain one or many scripts with .sh extension.\nEvery script wil be wrapped with 'elc --tag=hook' command.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			Pc = &RealPC{}

			hooksFolder := args[0]
			err := SetGitHooks(hooksFolder, os.Args[0])
			if err != nil {
				return err
			}
			return nil
		},
	}
	parentCommand.AddCommand(command)
}

func NewUpdateCommand(parentCommand *cobra.Command) {
	var version string
	var command = &cobra.Command{
		Use:   "update",
		Short: "Update elc binary",
		Long:  "Download new version of ELC, place it to /opt/elc/ and update symlink at /usr/local/bin.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			Pc = &RealPC{}

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
		},
	}
	command.Flags().StringVar(&version, "version", "", "desired version of elc")
	parentCommand.AddCommand(command)
}

func NewFixUpdateCommand(parentCommand *cobra.Command) {
	var version string
	var command = &cobra.Command{
		Use:   "fix-update-command",
		Short: "Set actual update command to ~/.elc.yaml",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			Pc = &RealPC{}

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
		},
	}
	command.Flags().StringVar(&version, "version", "", "desired version of elc")
	parentCommand.AddCommand(command)
}
