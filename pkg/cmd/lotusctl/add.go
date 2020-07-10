package lotusctl

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/shima-park/lotus/pkg/rpc/proto"
	"github.com/spf13/cobra"
)

func NewAddCmd(cmds ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add (RESOURCE/NAME | -f FILENAME)",
		Short: "Add a resource to the server",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(cmds...)

	return cmd
}

func NewAddExecutorCmd() *cobra.Command {
	var file string
	var _type string
	var name string
	var schedule string
	var bootstrap bool
	var processors []string
	var components []string
	cmd := &cobra.Command{
		Use:     "executor (PATH | URL | -n EXECUTOR_NAME -p PROCESSOR_NAME -c COMPONENT_NAME)",
		Aliases: []string{"exec"},
		Short:   "Add a executor to the server",
		Run: func(cmd *cobra.Command, args []string) {
			if file != "" {
				if len(args) == 0 {
					handleErr(errors.New("You need to provide a executor name."))
				}

				file, err := tryDownloadAndCheckPath(file)
				handleErr(err)

				data, err := ioutil.ReadFile(file)
				handleErr(err)

				err = newClient().AddExecutor(&proto.AddExecutorRequest{
					Config: data,
				})
				handleErr(err)
			} else if _type != "" && (len(processors) > 0 || len(components) > 0) {
				c := newClient()
				origin, err := c.GenerateExecutorConfig(
					&proto.GenerateExecutorConfigRequest{
						Name:       name,
						Type:       _type,
						Components: components,
						Processors: processors,
					},
				)
				handleErr(err)

				err = runEditor(
					[]byte(origin),
					func(edited []byte) error {
						fmt.Println("======", string(edited))
						err = newClient().PutExecutor(&proto.PutExecutorRequest{
							Config: edited,
						})
						return err
					})
				handleErr(err)
			} else {
				fmt.Println("-f executor.yaml or -n test -t pipeliner -p read_line -c es_client you at least provide one of them")
			}

		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to executor config")
	cmd.Flags().StringVarP(&_type, "type", "t", "", "type of executor")
	cmd.Flags().StringVarP(&name, "name", "n", "", "name of executor")
	cmd.Flags().StringVarP(&schedule, "schedule", "s", "", "name of executor")
	cmd.Flags().BoolVarP(&bootstrap, "bootstrap", "b", false, "whether to start with the server")
	cmd.Flags().StringSliceVarP(&processors, "processors", "p", nil, "processors of executor")
	cmd.Flags().StringSliceVarP(&components, "components", "c", nil, "components of executor")

	return cmd
}

func NewAddPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugin (PATH | URL)",
		Aliases: []string{"plug"},
		Short:   "Add a plugin to the server",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				handleErr(errors.New("You must provide a plugin path"))
			}

			var paths []string
			for _, path := range args {
				path, err := tryDownloadAndCheckPath(path)
				handleErr(err)

				paths = append(paths, path)
			}

			c := newClient()
			for _, path := range paths {
				data, err := ioutil.ReadFile(path)
				handleErr(err)

				err = c.AddPlugin(&proto.AddPluginRequest{
					Name:              filepath.Base(path),
					ShareObjectBinary: data,
				})
				handleErr(err)
			}
		},
	}
	return cmd
}

func init() {
	rootCmd.AddCommand(
		NewAddCmd(
			NewAddExecutorCmd(), NewAddPluginCmd(),
		),
	)
}
