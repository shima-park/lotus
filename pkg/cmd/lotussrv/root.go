package lotussrv

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/shima-park/lotus/pkg/rpc/http/server"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lotus",
	Short: "lotus is a pipeline-based task scheduling center",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

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

func Execute(version, branch, commit, built string) {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", version)
			fmt.Printf("Branch: %s\n", branch)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Built: %s\n", built)
		},
	})

	rootCmd.AddCommand(
		NewCmdServer(NewCmdServerRun(version, branch, commit, built)),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
