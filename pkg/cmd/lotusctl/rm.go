package lotusctl

import (
	"errors"

	"github.com/spf13/cobra"
)

func NewRMCmd(cmds ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm (RESOURCE/NAME | -f FILENAME)",
		Short: "Remove a resource on the server",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(cmds...)
	return cmd
}

func NewRMPipeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "executor (NAME)",
		Aliases: []string{"pipe"},
		Short:   "Remove a executor's config on the server",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				handleErr(errors.New("You must provide a executor name"))
			}
			err := newClient().Executor.Remove(args...)
			handleErr(err)
		},
	}
	return cmd
}

func NewRMPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugin (NAME)",
		Aliases: []string{"plug"},
		Short:   "Remove a plugin on the server",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				handleErr(errors.New("You must provide a plugin name"))
			}
			err := newClient().Plugin.Remove(args...)
			handleErr(err)
		},
	}
	return cmd
}

func init() {
	rootCmd.AddCommand(
		NewRMCmd(
			NewRMPipeCmd(), NewRMPluginCmd(),
		),
	)
}
