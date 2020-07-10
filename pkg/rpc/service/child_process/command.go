package child_process

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/docker/docker/pkg/reexec"
	"github.com/pkg/errors"
)

func RunCmd(cmdStr string, args []string, ret interface{}) error {
	cmd := reexec.Command(append([]string{cmdStr}, args...)...)
	var stdOut, errOut bytes.Buffer
	cmd.Stderr = &errOut
	cmd.Stdout = &stdOut
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, errOut.String())
	}
	fmt.Println("command:", cmdStr)
	fmt.Println("errOut:", errOut.String())
	fmt.Println("stdOut:", stdOut.String())
	if errOut.Len() > 0 {
		return errors.New(errOut.String())
	}

	if ret != nil {
		err = json.Unmarshal(bytes.TrimSpace(stdOut.Bytes()), ret)
		return err
	}
	return nil
}

func Success(data interface{}) {
	b, err := json.Marshal(data)
	Failed(err)

	fmt.Fprint(os.Stdout, string(b))
	os.Exit(0)
}

func Failed(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}
