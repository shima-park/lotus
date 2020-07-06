package lotusctl

import (
	"github.com/shima-park/lotus/pkg/rpc/proto"
	"github.com/spf13/cobra"
)

var cmdExecutor = &cobra.Command{
	Use:     "executor",
	Aliases: []string{"exec"},
	Short:   "Commands to control executor",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var cmdStartPipe = &cobra.Command{
	Use:   "start",
	Short: "start a executor",
	Run: func(cmd *cobra.Command, args []string) {
		err := newClient().Executor.Control(proto.ControlCommandStart, args...)
		handleErr(err)
	},
}

var cmdStopPipe = &cobra.Command{
	Use:   "stop",
	Short: "stop a executor",
	Run: func(cmd *cobra.Command, args []string) {
		err := newClient().Executor.Control(proto.ControlCommandStop, args...)
		handleErr(err)
	},
}

var cmdRestartPipe = &cobra.Command{
	Use:   "restart",
	Short: "restart a executor",
	Run: func(cmd *cobra.Command, args []string) {
		err := newClient().Executor.Control(proto.ControlCommandRestart, args...)
		handleErr(err)
	},
}

func init() {
	cmdExecutor.AddCommand(
		cmdStartPipe, cmdStopPipe, cmdRestartPipe,
	)
	rootCmd.AddCommand(cmdExecutor)
}
