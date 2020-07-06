package lotusctl

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

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

func NewEditPipeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "executor (NAME)",
		Aliases: []string{"exec"},
		Short:   "Edit a executor's config on the server",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				handleErr(errors.New("You must provide a executor name"))
			}
			c := newClient()
			exec, err := c.Executor.Find(args[0])
			handleErr(err)

			err = runEditor(
				exec.RawConfig,
				func(config []byte) error {
					return c.Executor.Recreate(exec.Name, config)
				},
				false)
			handleErr(err)
		},
	}
	return cmd
}

func runEditor(origin []byte, callback func(config []byte) error, isAdd bool) error {
	edit := editor.NewDefaultEditor([]string{"EDITOR"})
	buff := bytes.NewBuffer(origin)
	edited, path, err := edit.LaunchTempFile("edit-", "", buff)
	if err != nil {
		return err
	}
	defer os.Remove(path)

	if !isAdd && bytes.Equal(origin, edited) {
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
			NewEditPipeCmd(),
		),
	)
}
