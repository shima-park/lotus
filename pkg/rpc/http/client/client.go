package client

import (
	"bytes"
	"mime/multipart"
	stdhttp "net/http"

	"github.com/shima-park/lotus/pkg/rpc/proto"
	"github.com/shima-park/lotus/pkg/util/http"
)

type Client struct {
	apiBuilder
}

func NewClient(addr string) *Client {
	addr = http.NormalizeURL(addr)
	return &Client{apiBuilder{addr}}
}

type apiBuilder struct {
	addr string
}

func (b *apiBuilder) api(path string) string {
	return b.addr + path
}

func (c *Client) GenerateExecutorConfig(req *proto.GenerateExecutorConfigRequest) (string, error) {
	var s string
	err := http.PostJSON(c.api("/executor/generate-config"), req, &s)
	return s, err
}

func (c *Client) AddExecutor(req *proto.AddExecutorRequest) error {
	return http.PostJSON(c.api("/executor/add"), req, nil)
}

func (c *Client) PutExecutor(req *proto.PutExecutorRequest) error {
	return http.PostJSON(c.api("/executor/put"), req, nil)
}

func (c *Client) GetExecutor(req *proto.GetExecutorRequest) (*proto.GetExecutorResponse, error) {
	var resp proto.GetExecutorResponse
	err := http.GetJSON(c.api("/executor/get"), &resp)
	return &resp, err
}

func (c *Client) RemoveExecutor(req *proto.RemoveExecutorRequest) error {
	return http.PostJSON(c.api("/executor/remove"), req, nil)
}

func (c *Client) RegisterExecutorProcess(req *proto.RegisterExecutorProcessRequest) error {
	return http.PostJSON(c.api("/executor-process/register"), req, nil)
}

func (c *Client) RemoveExecutorProcess(req *proto.RemoveExecutorProcessRequest) error {
	return http.PostJSON(c.api("/executor-process/remove"), req, nil)
}

func (c *Client) ListExecutor() ([]proto.ExecutorView, error) {
	var res []proto.ExecutorView
	err := http.GetJSON(c.api("/executor/list"), &res)
	return res, err
}

func (c *Client) ListExecutorProcess() ([]proto.ExecutorProcessView, error) {
	var res []proto.ExecutorProcessView
	err := http.GetJSON(c.api("/executor-process/list"), &res)
	return res, err
}

func (c *Client) ListComponent() ([]proto.ComponentView, error) {
	var res []proto.ComponentView
	err := http.GetJSON(c.api("/component/list"), &res)
	return res, err
}

func (c *Client) ListProcessor() ([]proto.ProcessorView, error) {
	var res []proto.ProcessorView
	err := http.GetJSON(c.api("/processor/list"), &res)
	return res, err
}

func (c *Client) ListPlugin() ([]proto.PluginView, error) {
	var res []proto.PluginView
	err := http.GetJSON(c.api("/plugin/list"), &res)
	return res, err
}

func (c *Client) AddPlugin(req *proto.AddPluginRequest) error {
	return c.uploadPlugin("/plugin/add", req.Name, req.ShareObjectBinary)
}

func (c *Client) PutPlugin(req *proto.PutPluginRequest) error {
	return c.uploadPlugin("/plugin/put", req.Name, req.ShareObjectBinary)
}

func (c *Client) uploadPlugin(api, name string, bin []byte) error {
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

	request, err := stdhttp.NewRequest("POST", c.api(api), body)
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

func (c *Client) RemovePlugin(req *proto.RemovePluginRequest) error {
	return http.PostJSON(c.api("/plugin/remove"), req, nil)
}
