package cmd

import (
	"github.com/ensi-platform/elc/actions"
	"github.com/ensi-platform/elc/core"
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
	cmd.Flags().BoolVar(&globalOptions.NoTty, "no-tty", false, "disable pseudo-TTY allocation")
}

func InitCobra() *cobra.Command {
	globalOptions = core.GlobalOptions{}
	var rootCmd = &cobra.Command{
		Use:          "elc [command]",
		Args:         cobra.MinimumNArgs(0),
		SilenceUsage: true,
		//SilenceErrors: true,
		Version: core.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			core.Pc = &core.RealPC{}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			} else {
				globalOptions.Cmd = args
				return actions.ExecAction(&globalOptions)
			}
		},
	}

	rootCmd.Flags().SetInterspersed(false)

	rootCmd.PersistentFlags().StringVarP(&globalOptions.ComponentName, "component", "c", "", "name of component")
	rootCmd.PersistentFlags().StringVarP(&globalOptions.WorkspaceName, "workspace", "w", "", "name of workspace")
	rootCmd.PersistentFlags().StringVar(&globalOptions.ComponentName, "svc", "", "name of current component (deprecated, alias for component)")
	rootCmd.PersistentFlags().BoolVar(&globalOptions.Debug, "debug", false, "print debug messages")
	rootCmd.PersistentFlags().BoolVar(&globalOptions.DryRun, "dry-run", false, "do not execute real command, only debug")
	rootCmd.PersistentFlags().StringVar(&globalOptions.Tag, "tag", "", "select all components with tag")
	rootCmd.PersistentFlags().StringVarP(&globalOptions.Branch, "branch", "b", "", "use component worktree for the given git branch")
	rootCmd.PersistentFlags().StringVar(&globalOptions.Source, "source", "", "create missing branch from this ref (branch name or HEAD)")

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
	NewServiceLaunchCommand(rootCmd)
	NewServiceExecCommand(rootCmd)
	NewServiceRunCommand(rootCmd)
	NewServiceSetHooksCommand(rootCmd)
	NewServiceCloneCommand(rootCmd)
	NewServiceRunHookCommand(rootCmd)
	NewServiceListCommand(rootCmd)
	NewWorktreeCommand(rootCmd)

	return rootCmd
}

func NewWorkspaceCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:     "workspace",
		Aliases: []string{"ws"},
	}
	NewWorkspaceListCommand(command)
	NewWorkspaceAddCommand(command)
	NewWorkspaceRemoveCommand(command)
	NewWorkspaceShowCommand(command)
	NewWorkspaceSelectCommand(command)
	NewWorkspaceSetRootCommand(command)
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
		Use:   "add [NAME] [PATH]",
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

func NewWorkspaceRemoveCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "remove [NAME]",
		Short: "Remove workspace from ~/.elc.yaml",
		Long:  "Remove workspace from ~/.elc.yaml.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			return actions.RemoveWorkspaceAction(name)
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
			return actions.ShowCurrentWorkspaceAction(&globalOptions)
		},
	}
	parentCommand.AddCommand(command)
}

func NewWorkspaceSelectCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "select [NAME]",
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

func NewWorkspaceSetRootCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "set-root [NAME] [PATH]",
		Short: "Set root path for workspace",
		Long:  "Set root path for workspace.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return actions.SetRootPathAction(args[0], args[1])
		},
	}
	parentCommand.AddCommand(command)
}

func NewServiceStartCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "start [OPTIONS] [NAME]",
		Short: "Start one or more services",
		Long:  "Start one or more services.\nBy default starts service found with current directory, but you can pass one or more service names instead.",
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
		Use:   "stop [OPTIONS] [NAME]",
		Short: "Stop one or more services",
		Long:  "Stop one or more services.\nBy default stops service found with current directory, but you can pass one or more service names instead.",
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
		Use:   "destroy [OPTIONS] [NAME]",
		Short: "Stop and remove containers of one or more services",
		Long:  "Stop and remove containers of one or more services.\nBy default destroys service found with current directory, but you can pass one or more service names instead.",
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
		Use:   "restart [OPTIONS] [NAME]",
		Short: "Restart one or more services",
		Long:  "Restart one or more services.\nBy default restart service found with current directory, but you can pass one or more service names instead.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return actions.RestartServiceAction(hardRestart, args, &globalOptions)
		},
	}
	command.Flags().BoolVar(&hardRestart, "hard", false, "destroy container instead of stop it before start")
	parentCommand.AddCommand(command)
}

func NewServiceVarsCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "vars [NAME]",
		Short: "Print all variables computed for service",
		Long:  "Print all variables computed for service.\nBy default uses service found with current directory, but you can pass name of another service instead.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return actions.PrintVarsAction(&globalOptions, args)
		},
	}
	parentCommand.AddCommand(command)
}

func NewServiceComposeCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "compose [OPTIONS] [COMMAND]",
		Short: "Run docker-compose command",
		Long:  "Run docker-compose command.\nBy default uses service found with current directory.",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			} else {
				return actions.ComposeCommandAction(&globalOptions, args)
			}
		},
	}
	command.Flags().SetInterspersed(false)
	parentCommand.AddCommand(command)
}

func NewServiceWrapCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "wrap [COMMAND]",
		Short: "Execute command on host with env variables for service. For module uses variables of linked service",
		Long:  "Execute command on host with env variables for service.\nFor module uses variables of linked service.\nBy default uses service/module found with current directory.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return actions.WrapCommandAction(&globalOptions, args)
		},
	}
	parentCommand.AddCommand(command)
}

func NewServiceLaunchCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "launch [COMMAND]",
		Short: "Run command on host in component directory",
		Long:  "Run command on host in the component folder (main clone or worktree).\nComponent is selected via CWD, -c/--component.\nIf --branch is set (or CWD is a worktree), the worktree is used; missing worktrees are created automatically (fetch + remote branch resolve).",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return actions.LaunchAction(&globalOptions, args)
		},
	}
	command.Flags().SetInterspersed(false)
	parentCommand.AddCommand(command)
}

func NewServiceExecCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "exec [OPTIONS] [COMMAND]",
		Short: "Execute command in container. For module uses container of linked service",
		Long:  "Execute command in container. For module uses container of linked service.\nBy default uses service/module found with current directory. Starts service if it is not running.",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Usage()
			} else {
				globalOptions.Cmd = args
				return actions.ExecAction(&globalOptions)
			}
		},
	}
	command.Flags().SetInterspersed(false)
	parseStartFlags(command)
	parseExecFlags(command)
	parentCommand.AddCommand(command)
}
func NewServiceRunCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "run [OPTIONS] [COMMAND]",
		Short: "Run new container with command. For module uses container of linked service",
		Long:  "Run new container with command. For module uses container of linked service.\nBy default uses service/module found with current directory. Starts service if it is not running.",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Usage()
			} else {
				globalOptions.Cmd = args
				return actions.RunAction(&globalOptions)
			}
		},
	}
	command.Flags().SetInterspersed(false)
	parseExecFlags(command)
	parentCommand.AddCommand(command)
}

func NewServiceSetHooksCommand(parentCommand *cobra.Command) {
	var native bool
	var command = &cobra.Command{
		Use:   "set-hooks [HOOKS_DIR]",
		Short: "Install hooks from specified folder to .git/hooks",
		Long:  "Install hooks from specified folder to .git/hooks.\nHOOKS_PATH must contain subdirectories with names as git hooks, eg. 'pre-commit'.\nOne subdirectory can contain one or many scripts with .sh extension.\nEvery script will be wrapped with 'elc --tag=hook' command.\nWith --native, sets git core.hooksPath to HOOKS_DIR instead of generating wrappers in .git/hooks.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return actions.SetGitHooksAction(&globalOptions, args[0], os.Args[0], native)
		},
	}
	command.Flags().BoolVar(&native, "native", false, "set core.hooksPath to HOOKS_DIR instead of generating .git/hooks wrappers")
	parentCommand.AddCommand(command)
}

func NewServiceCloneCommand(parentCommand *cobra.Command) {
	var noHook bool
	var command = &cobra.Command{
		Use:           "clone [NAME]",
		Short:         "Clone component to its path",
		Long:          "Clone component to its path.",
		SilenceUsage:  false,
		SilenceErrors: false,
		Args:          cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return actions.CloneComponentAction(&globalOptions, args, noHook)
		},
	}

	command.Flags().BoolVar(&noHook, "no-hook", false, "do not execute hook script after cloning")
	parentCommand.AddCommand(command)
}

func NewServiceRunHookCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:     "run-hook [HOOK] [-- ARGS...]",
		Aliases: []string{"rh"},
		Short:   "Run a component hook from workspace.yaml",
		Long:    "Run a named component hook defined in workspace.yaml (hooks.after_clone, hooks.worktree_create, hooks.worktree_remove).\nBy default uses the component found with current directory.\nArguments after -- are passed to the hook script.",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return actions.RunHookAction(&globalOptions, args[0], args[1:])
		},
	}
	command.Flags().SetInterspersed(false)
	parentCommand.AddCommand(command)
}

func NewServiceListCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:   "list [OPTIONS]",
		Short: "Show list of services",
		Long:  "Show list of services.\nCan be used in scripts for loops.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return actions.ListServicesAction(&globalOptions)
		},
	}
	parentCommand.AddCommand(command)
}

func NewWorktreeCommand(parentCommand *cobra.Command) {
	var noHook bool
	var command = &cobra.Command{
		Use:     "worktree [branch]",
		Aliases: []string{"wt"},
		Short:   "Create a git worktree instance of a component",
		Long:    "Create a git worktree of the current or selected component at $WORKTREES_PATH/<component>/<branch>.\nFetches remotes and creates a local tracking branch when the branch exists only on origin.\nIf the branch does not exist anywhere, fails unless --source=<branch|HEAD> is set.\nAfter creation, optional hooks.worktree_create is executed.\nArguments after -- are passed to the hook script.",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return actions.AddWorktreeAction(&globalOptions, args[0], args[1:], noHook)
		},
	}
	command.Flags().BoolVar(&noHook, "no-hook", false, "do not execute hooks.worktree_create")
	NewWorktreeAddCommand(command, &noHook)
	NewWorktreeListCommand(command)
	NewWorktreeRemoveCommand(command)
	parentCommand.AddCommand(command)
}

func NewWorktreeAddCommand(parentCommand *cobra.Command, noHook *bool) {
	var command = &cobra.Command{
		Use:   "add [branch]",
		Short: "Create a git worktree instance of a component",
		Long:  "Create a git worktree of the current or selected component at $WORKTREES_PATH/<component>/<branch>.\nArguments after -- are passed to hooks.worktree_create.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return actions.AddWorktreeAction(&globalOptions, args[0], args[1:], *noHook)
		},
	}
	command.Flags().BoolVar(noHook, "no-hook", false, "do not execute hooks.worktree_create")
	parentCommand.AddCommand(command)
}

func NewWorktreeListCommand(parentCommand *cobra.Command) {
	var command = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List component worktree instances",
		Long:    "List worktrees under $WORKTREES_PATH.\nOutput lines: <component>\\t<branch>\\t<path>.\nUse -c/--component to filter by component.",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return actions.ListWorktreesAction(&globalOptions)
		},
	}
	parentCommand.AddCommand(command)
}

func NewWorktreeRemoveCommand(parentCommand *cobra.Command) {
	var force bool
	var noHook bool
	var command = &cobra.Command{
		Use:     "remove [branch]",
		Aliases: []string{"rm"},
		Short:   "Remove a component worktree instance",
		Long:    "Run optional hooks.worktree_remove, destroy containers, and remove the git worktree.\nBranch can be taken from the current directory, argument, or --branch.",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branch := ""
			if len(args) > 0 {
				branch = args[0]
			}
			return actions.RemoveWorktreeAction(&globalOptions, branch, force, noHook)
		},
	}
	command.Flags().BoolVar(&force, "force", false, "force git worktree removal and ignore compose/hook errors")
	command.Flags().BoolVar(&noHook, "no-hook", false, "do not execute hooks.worktree_remove")
	parentCommand.AddCommand(command)
}
