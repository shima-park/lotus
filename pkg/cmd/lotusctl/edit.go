package lotusctl

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/shima-park/lotus/pkg/rpc/proto"
	"github.com/shima-park/lotus/pkg/util/editor"
	"github.com/spf13/cobra"
)

func NewEditCmd(cmds ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit (RESOURCE/NAME | -f FILENAME)",
		Short: "Edit a resource on the server",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(cmds...)
	return cmd
}

func NewEditExecCmd() *cobra.Command {
	var _type string
	cmd := &cobra.Command{
		Use:     "executor (NAME)",
		Aliases: []string{"exec"},
		Short:   "Edit a executor's config on the server",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				handleErr(errors.New("You must provide a executor name"))
			}
			c := newClient()
			resp, err := c.GetExecutor(&proto.GetExecutorRequest{
				Name: args[0],
			})
			handleErr(err)

			err = runEditor(
				resp.Config,
				func(edited []byte) error {
					err := c.PutExecutor(&proto.PutExecutorRequest{
						Config: edited,
					})
					return err
				})
			handleErr(err)
		},
	}
	cmd.Flags().StringVarP(&_type, "type", "t", "", "type of executor")
	return cmd
}

func runEditor(origin []byte, callback func(config []byte) error) error {
	edit := editor.NewDefaultEditor([]string{"EDITOR"})
	buff := bytes.NewBuffer(origin)
	edited, path, err := edit.LaunchTempFile("edit-", "", buff)
	if err != nil {
		return err
	}
	defer os.Remove(path)

	if bytes.Equal(origin, edited) {
		fmt.Println("Edit cancelled, no changes made.")
		return nil
	}

	lines, err := hasLines(bytes.NewBuffer(edited))
	handleErr(err)
	if !lines {
		fmt.Println("Edit cancelled, saved file was empty.")
		return nil
	}

	return callback(edited)
}

func hasLines(r io.Reader) (bool, error) {
	// TODO: if any files we read have > 64KB lines, we'll need to switch to bytes.ReadLine
	// TODO: probably going to be secrets
	s := bufio.NewScanner(r)
	for s.Scan() {
		if line := strings.TrimSpace(s.Text()); len(line) > 0 && line[0] != '#' {
			return true, nil
		}
	}
	if err := s.Err(); err != nil && err != io.EOF {
		return false, err
	}
	return false, nil
}

func init() {
	rootCmd.AddCommand(
		NewEditCmd(
			NewEditExecCmd(),
		),
	)
}
