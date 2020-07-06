package client

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	stdhttp "net/http"
	"os"

	"github.com/shima-park/lotus/pkg/rpc/proto"
	"github.com/shima-park/lotus/pkg/util/http"
)

var _ proto.Plugin = &plugin{}

type plugin struct {
	apiBuilder
}

func (c *plugin) List() ([]proto.PluginView, error) {
	var res []proto.PluginView
	err := http.GetJSON(c.api("/plugin/list"), &res)
	return res, err
}

func (c *plugin) AddPath(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	fi, err := file.Stat()
	if err != nil {
		return err
	}

	return c.Add(fi.Name(), data)
}

func (c *plugin) Add(name string, bin []byte) error {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("plugin", name)
	if err != nil {
		return err
	}
	_, _ = part.Write(bin)

	err = writer.Close()
	if err != nil {
		return err
	}

	request, err := stdhttp.NewRequest("POST", c.api("/plugin/upload"), body)
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())

	client := &stdhttp.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = http.HandleBody(resp.Body, nil)
	if err != nil {
		return err
	}
	return nil
}

func (p *plugin) Remove(names ...string) error {
	return http.PostJSON(p.api("/plugin/remove"), &names, nil)
}
