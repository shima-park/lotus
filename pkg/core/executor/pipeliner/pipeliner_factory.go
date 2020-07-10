package pipeliner

import (
	"github.com/shima-park/seed/executor"
	"gopkg.in/yaml.v2"
)

var (
	factory      executor.Factory  = &PipelinerFactory{}
	_            executor.Executor = &pipeliner{}
	sampleConfig                   = `
name: test
kind: pipeliner
schedule: ""  # 为空时死循环调度，支持cron表达式
circuit_breaker_samples: 10 # 熔断采样数量
circuit_breaker_rate: 0.6 # 熔断采样率
bootstrap: true # 是否随进程启动而启动
components: # 组件配置列表
#  - test_component: |
#    name: test
processors: # 处理器配置列表
#  - test_component: |
#    name: test
stream: # 处理器执行顺序，n叉树结构
#  name: test       # 需要执行的processor的名字
#  replica: 3       # 该processor起多少个goroutine执行
#  buffer_size: 10  # 向下输出结果集channel的缓冲区多大
#  childs:          # 按stream结构向下继续定义流程
`
)

func init() {
	if err := executor.Register("pipeliner", factory); err != nil {
		panic(err)
	}
}

type PipelinerFactory struct {
}

func (f *PipelinerFactory) SampleConfig() string {
	return sampleConfig
}

func (f *PipelinerFactory) Description() string {
	return "Pipeliner is a pipeline-based task executor"
}

func (f *PipelinerFactory) New(config string) (executor.Executor, error) {
	var conf Config
	err := yaml.Unmarshal([]byte(config), &conf)
	if err != nil {
		return nil, err
	}
	p := NewPipelineByConfig(conf)
	return p, nil
}
