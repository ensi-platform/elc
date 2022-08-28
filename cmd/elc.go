package cmd

import (
	"github.com/madridianfox/elc/actions"
	"github.com/madridianfox/elc/core"
	"github.com/spf13/cobra"
	"os"
)

var globalOptions core.GlobalOptions

func parseStartFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&globalOptions.Force, "force", false, "force start dependencies, even if service already started")
	cmd.Flags().StringVar(&globalOptions.Mode, "mode", "default", "start only dependencies with specified mode, by default starts 'default' dependencies")
}

func parseExecFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&globalOptions.UID, "uid", -1, "use another uid, by default uses uid of current user")
}

func InitCobra() *cobra.Command {
	globalOptions = core.GlobalOptions{}
	var rootCmd = &cobra.Command{
		Use:           "elc",
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       core.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			core.Pc = &core.RealPC{}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			globalOptions.Cmd = args
			return actions.ExecAction(globalOptions)
		},
	}

	rootCmd.Flags().SetInterspersed(false)

	rootCmd.PersistentFlags().StringVarP(&globalOptions.ComponentName, "component", "c", "", "name of component")
	rootCmd.PersistentFlags().StringVarP(&globalOptions.WorkspaceName, "workspace", "w", "", "name of workspace")
	rootCmd.PersistentFlags().StringVar(&globalOptions.ComponentName, "svc", "", "name of current component (deprecated, alias for component)")
	rootCmd.PersistentFlags().BoolVar(&globalOptions.Debug, "debug", false, "print debug messages")

	parseStartFlags(rootCmd)
	parseExecFlags(rootCmd)

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
		Use:     "workspace",
		Aliases: []string{"ws"},
	}
	NewWorkspaceListCommand(command)
	NewWorkspaceAddCommand(command)
	NewWorkspaceShowCommand(command)
	NewWorkspaceSelectCommand(command)
	parentCommand.AddCommand(command)
}

func NewWorkspaceListCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Show list of registered workspaces",
		Long:    "Show list of registered workspaces.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return actions.ListWorkspacesAction()
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
			name := args[0]
			wsPath := args[1]

			return actions.AddWorkspaceAction(name, wsPath)
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
			core.Pc = &core.RealPC{}
			return actions.ShowCurrentWorkspaceAction()
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
			name := args[0]
			return actions.SelectWorkspaceAction(name)
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
			return actions.StartServiceAction(&globalOptions, args)
		},
	}
	parseStartFlags(command)
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
			return actions.StopServiceAction(stopAll, args, false, &globalOptions)
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
			return actions.StopServiceAction(destroyAll, args, true, &globalOptions)
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
			return actions.RestartServiceAction(hardRestart, args, &globalOptions)
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
			return actions.PrintVarsAction(args)
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
			return actions.ComposeCommandAction(args, globalOptions)
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
			return actions.WrapCommandAction(globalOptions, args)
		},
	}
	parentCommand.AddCommand(command)
}

func NewServiceExecCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "exec",
		Short: "Execute command in container. For module uses container of linked service",
		Long:  "By default uses service/module found with current directory. Starts service if it is not running.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			globalOptions.Cmd = args
			return actions.ExecAction(globalOptions)
		},
	}
	command.Flags().SetInterspersed(false)
	parseStartFlags(command)
	parseExecFlags(command)
	parentCommand.AddCommand(command)
}

func NewServiceSetHooksCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "set-hooks",
		Short: "Install hooks from specified folder to .git/hooks",
		Long:  "HOOKS_PATH must contain subdirectories with names as git hooks, eg. 'pre-commit'.\nOne subdirectory can contain one or many scripts with .sh extension.\nEvery script wil be wrapped with 'elc --tag=hook' command.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return actions.SetGitHooksAction(args[0], os.Args[0])
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
			return actions.UpdateBinaryAction(version)
		},
	}
	command.Flags().StringVar(&version, "version", "", "desired version of elc")
	parentCommand.AddCommand(command)
}

func NewFixUpdateCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "fix-update-command",
		Short: "Set actual update command to ~/.elc.yaml",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return actions.FixUpdateBinaryCommandAction()
		},
	}
	parentCommand.AddCommand(command)
}
