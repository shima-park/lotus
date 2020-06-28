package client

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/shima-park/lotus/pkg/rpc/proto"
)

var _ proto.Plugin = &plugin{}

type plugin struct {
	apiBuilder
}

func (c *plugin) List() ([]proto.PluginView, error) {
	var res []proto.PluginView
	err := GetJSON(c.api("/plugin/list"), &res)
	return res, err
}

func (c *plugin) Open(path string) error {
	req := &proto.PluginOpenRequest{
		Path: path,
	}
	return PostJSON(c.api("/plugin/open"), req, nil)
}

func (c *plugin) Add(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	fi, err := file.Stat()
	if err != nil {
		return err
	}
	file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("plugin", fi.Name())
	if err != nil {
		return err
	}
	_, _ = part.Write(fileContents)

	err = writer.Close()
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", c.api("/plugin/upload"), body)
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = handleBody(resp.Body, nil)
	if err != nil {
		return err
	}
	return nil
}

func (p *plugin) Remove(names ...string) error {
	return PostJSON(p.api("/plugin/remove"), &names, nil)
}
