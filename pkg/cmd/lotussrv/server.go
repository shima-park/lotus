package lotussrv

import (
	"os"
	"os/signal"

	"github.com/shima-park/lotus/pkg/rpc/server"
	"github.com/spf13/cobra"
)

func NewCmdServer(cmds ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"serv", "srv"},
		Short:   "Commands to control server",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	cmd.AddCommand(cmds...)

	return cmd
}

func NewCmdServerRun(version, branch, commit, built string) *cobra.Command {
	var metaPath string
	var httpAddr string
	cmd := &cobra.Command{
		Use:   "run",
		Short: "run a lotus server",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := server.New(
				server.HTTPAddr(httpAddr),
				server.MetadataPath(metaPath),
				server.Version(version),
				server.Branch(branch),
				server.Commit(commit),
				server.Built(built),
			)
			if err != nil {
				panic(err)
			}

			if err := c.Serve(); err != nil {
				panic(err)
			}

			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt)
			<-signals

			c.Stop()
		},
	}
	cmd.Flags().StringVar(&metaPath, "meta", "", "path to metadata")
	cmd.Flags().StringVar(&httpAddr, "http", "", "listen on address")
	return cmd
}
