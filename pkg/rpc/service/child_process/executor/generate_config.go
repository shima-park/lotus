package executor

import (
	"errors"
	"flag"
	"os"
	"strings"

	_ "github.com/shima-park/lotus/pkg/core/component/include"
	"github.com/shima-park/lotus/pkg/core/executor/pipeliner"

	"github.com/docker/docker/pkg/reexec"
	"github.com/shima-park/lotus/pkg/core/common/inject"
	cp "github.com/shima-park/lotus/pkg/rpc/service/child_process"
	"github.com/shima-park/seed/component"
	"github.com/shima-park/seed/plugin"
	"github.com/shima-park/seed/processor"
	"gopkg.in/yaml.v2"
)

const (
	CHILD_PROCESS_GENERATE_EXECUTOR_CONFIG = "generate-executor-config"
)

func init() {
	reexec.Register(CHILD_PROCESS_GENERATE_EXECUTOR_CONFIG, generateExecutorConfig)
	if reexec.Init() {
		os.Exit(0)
	}
}

func generateExecutorConfig() {
	var _type = flag.String("type", "", "executor type")
	var components = flag.String("components", "", "executor's component list")
	var processors = flag.String("processors", "", "executor's processor list")
	var plugins = flag.String("plugins", "", "executor's plugins list")

	flag.Parse()

	for _, path := range strings.Split(*plugins, ",") {
		if strings.TrimSpace(path) == "" {
			continue
		}
		_, err := plugin.LoadPlugins(path)
		cp.Failed(err)
	}

	if *_type != "pipeliner" {
		cp.Failed(errors.New("Unsupported type " + *_type))
	}

	dependencyMap := map[string][]string{} // key:type value:inject_name
	processorList := strings.Split(*processors, ",")
	var processorConfigs []map[string]string
	streamConfig := &pipeliner.StreamConfig{}
	t := streamConfig
	for i, name := range processorList {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		f, err := processor.GetFactory(name)
		cp.Failed(err)

		t.Name = name
		if i != len(processorList)-1 { // 防止加上最后一个空childs
			t.Childs = []pipeliner.StreamConfig{
				pipeliner.StreamConfig{},
			}
			t = &t.Childs[0]
		}

		// 获取processor的component依赖项的type和injectName
		reqs, resps := inject.GetFuncReqAndRespReceptorList(f.Example())
		for _, r := range append(reqs, resps...) {
			dependencyMap[r.ReflectType] = append(dependencyMap[r.ReflectType], r.InjectName)
		}

		processorConfigs = append(processorConfigs, map[string]string{
			name: f.SampleConfig(),
		})
	}

	dependencyUsedMap := map[string]int{} // key:type value:index
	componentList := strings.Split(*components, ",")
	var componentConfigs []map[string]string
	for _, name := range componentList {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		f, err := component.GetFactory(name)
		cp.Failed(err)

		typeStr := f.ExampleType().String()

		// 设置component的注入名字和processor一致
		config := f.SampleConfig()
		if injectNames, ok := dependencyMap[typeStr]; ok {
			i := dependencyUsedMap[typeStr]
			var injectName string
			if i < len(injectNames) {
				injectName = injectNames[i]
			} else {
				injectName = injectNames[len(injectNames)-1]
			}
			config = setInjectName(injectName, config)
			dependencyUsedMap[typeStr]++
		}

		componentConfigs = append(componentConfigs, map[string]string{
			name: config,
		})
	}

	conf := &pipeliner.Config{
		Kind:       "pipeliner",
		Components: componentConfigs,
		Processors: processorConfigs,
		Stream:     *streamConfig,
	}

	b, err := yaml.Marshal(conf)
	cp.Failed(err)

	cp.Success(generateExecutorConfigResponse{
		Config: string(b),
	})
}

func setInjectName(name string, config string) string {
	m := map[string]interface{}{}
	_ = yaml.Unmarshal([]byte(config), &m)
	if _, ok := m["name"]; ok {
		m["name"] = name
	} else {
		return config
	}
	b, _ := yaml.Marshal(m)
	return string(b)
}

type generateExecutorConfigResponse struct {
	Config string `json:"config"`
}

func GenerateExecutorConfig(_type string, components, processors, plugins []string) (string, error) {
	var resp generateExecutorConfigResponse
	err := cp.RunCmd(
		CHILD_PROCESS_GENERATE_EXECUTOR_CONFIG,
		[]string{
			"--type", _type,
			"--components", strings.Join(components, ","),
			"--processors", strings.Join(processors, ","),
			"--plugins", strings.Join(plugins, ","),
		},
		&resp,
	)

	return resp.Config, err
}
