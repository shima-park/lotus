package service

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/docker/docker/pkg/reexec"
	"github.com/shima-park/lotus/pkg/common/plugin"
)

func init() {
	reexec.Register("test_plugin", testPlugin)
	if reexec.Init() {
		os.Exit(0)
	}
}

func testPlugin() {
	if len(os.Args) == 1 {
		return
	}

	paths := os.Args[1:]
	for _, path := range paths {
		err := plugin.LoadPlugins(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			break
		}
	}
}

func TestPlugin(paths ...string) error {
	cmd := reexec.Command(append([]string{"test_plugin"}, paths...)...)
	var errOut bytes.Buffer
	cmd.Stderr = &errOut
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return err
	}
	if errOut.Len() > 0 {
		return errors.New(errOut.String())
	}
	return nil
}
