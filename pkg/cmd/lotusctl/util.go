package lotusctl

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/shima-park/lotus/pkg/rpc/http/client"
)

func newClient(hosts ...string) *client.Client {
	for _, host := range append(
		hosts,
		os.Getenv("LOTUS_SERVER_ENV"),
		"localhost:8080",
	) {
		if host != "" {
			return client.NewClient(host)
		}
	}

	return nil
}

func renderTable(header []string, rows [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}

func handleErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func stringInSlice(t string, ss []string) bool {
	for _, s := range ss {
		if t == s {
			return true
		}
	}
	return false
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func tryDownloadAndCheckPath(path string) (string, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		f, err := ioutil.TempFile(os.TempDir(), "*_"+filepath.Base(path))
		if err != nil {
			return "", err
		}

		err = downloadFile(f.Name(), path)
		if err != nil {
			return "", err
		}

		path = f.Name()
	}

	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "", err
	}

	return path, err
}
