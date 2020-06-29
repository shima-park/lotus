package lotusctl

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/shima-park/lotus/pkg/pipeline"
	"github.com/shima-park/lotus/pkg/rpc/proto"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func NewAddCmd(cmds ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add (RESOURCE/NAME | -f FILENAME)",
		Short: "add a resource to the server",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(cmds...)
	return cmd
}

func NewAddPipelineCmd() *cobra.Command {
	var file string
	var name string
	var schedule string
	var bootstrap bool
	var processors []string
	var components []string
	cmd := &cobra.Command{
		Use:     "pipeline (PATH | URL | -n PIPELINE_NAME -p PROCESSOR_NAME -c COMPONENT_NAME)",
		Aliases: []string{"pipe"},
		Short:   "Add a pipeline to the server",
		Run: func(cmd *cobra.Command, args []string) {
			if file != "" {
				file, err := tryDownloadAndCheckPath(file)
				handleErr(err)

				data, err := ioutil.ReadFile(file)
				handleErr(err)

				var conf pipeline.Config
				err = yaml.Unmarshal(data, &conf)
				handleErr(err)

				err = newClient().Pipeline.Add(conf)
				handleErr(err)
			} else if name != "" || len(processors) > 0 || len(components) > 0 {
				c := newClient()
				conf, err := c.Pipeline.GenerateConfig(
					name,
					proto.WithSchedule(schedule),
					proto.WithBootstrap(bootstrap),
					proto.WithComponents(components),
					proto.WithProcessor(processors),
				)
				handleErr(err)

				origin, err := yaml.Marshal(conf)
				handleErr(err)

				err = runEditor(origin, c.Pipeline.Add, true)
				handleErr(err)
			} else {
				fmt.Println("-f pipeline.yaml or -n test -p read_line -c es_client you at least provide one of them")
			}

		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to pipeline config")
	cmd.Flags().StringVarP(&name, "name", "n", "", "name of pipeline")
	cmd.Flags().StringVarP(&schedule, "schedule", "s", "", "name of pipeline")
	cmd.Flags().BoolVarP(&bootstrap, "bootstrap", "b", false, "whether to start with the server")
	cmd.Flags().StringSliceVarP(&processors, "processors", "p", nil, "processors of pipeline")
	cmd.Flags().StringSliceVarP(&components, "components", "c", nil, "components of pipeline")

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
				err := c.Plugin.Add(path)
				handleErr(err)
			}
		},
	}
	return cmd
}

func init() {
	rootCmd.AddCommand(
		NewAddCmd(
			NewAddPipelineCmd(), NewAddPluginCmd(),
		),
	)
}
