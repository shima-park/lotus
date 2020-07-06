package service

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/shima-park/lotus/pkg/common/log"
	"github.com/shima-park/lotus/pkg/executor"
	"github.com/shima-park/lotus/pkg/rpc/proto"
)

type executorService struct {
	metadata  proto.Metadata
	rwlock    sync.RWMutex
	executors map[string]Executor // key: name value: Executor
}

type Executor struct {
	executor.Executor
	Type       string
	ConfigPath string
}

func NewExecutorService(metadata proto.Metadata) proto.Executor {
	return &executorService{
		metadata:  metadata,
		executors: map[string]Executor{},
	}
}

func (s *executorService) GenerateConfig(name string) (string, error) {
	//options := proto.NewConfigOptions(opts...)
	//
	//dependencyMap := map[string][]string{} // key:type value:inject_name
	//var processorConfigs []map[string]string
	//streamConfig := &executor.StreamConfig{}
	//t := streamConfig
	//for i, name := range options.Processors {
	//	name = strings.TrimSpace(name)
	//	f, err := processor.GetFactory(name)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	t.Name = name
	//	if i != len(options.Processors)-1 { // 防止加上最后一个空childs
	//		t.Childs = []executor.StreamConfig{
	//			executor.StreamConfig{},
	//		}
	//		t = &t.Childs[0]
	//	}
	//
	//	// 获取processor的component依赖项的type和injectName
	//	reqs, resps := getFuncReqAndRespReceptorList(f.Example())
	//	for _, r := range append(reqs, resps...) {
	//		dependencyMap[r.ReflectType] = append(dependencyMap[r.ReflectType], r.InjectName)
	//	}
	//
	//	processorConfigs = append(processorConfigs, map[string]string{
	//		name: f.SampleConfig(),
	//	})
	//}
	//
	//dependencyUsedMap := map[string]int{} // key:type value:index
	//var componentConfigs []map[string]string
	//for _, name := range options.Components {
	//	name = strings.TrimSpace(name)
	//	f, err := component.GetFactory(name)
	//	if err != nil {
	//		return nil, err
	//	}
	//	typeStr := f.ExampleType().String()
	//
	//	// 设置component的注入名字和processor一致
	//	config := f.SampleConfig()
	//	if injectNames, ok := dependencyMap[typeStr]; ok {
	//		i := dependencyUsedMap[typeStr]
	//		var injectName string
	//		if i < len(injectNames) {
	//			injectName = injectNames[i]
	//		} else {
	//			injectName = injectNames[len(injectNames)-1]
	//		}
	//		config = setInjectName(injectName, config)
	//		dependencyUsedMap[typeStr]++
	//	}
	//
	//	componentConfigs = append(componentConfigs, map[string]string{
	//		name: config,
	//	})
	//}
	//
	//conf := &executor.Config{
	//	Name:                  name,
	//	Schedule:              options.Schedule,
	//	CircuitBreakerSamples: options.CircuitBreakerSamples,
	//	CircuitBreakerRate:    options.CircuitBreakerRate,
	//	Bootstrap:             options.Bootstrap,
	//	Components:            componentConfigs,
	//	Processors:            processorConfigs,
	//	Stream:                *streamConfig,
	//}
	//
	//b, err := yaml.Marshal(conf)
	//return string(b), err
	return "", nil
}

func (s *executorService) Add(_type string, config []byte) error {
	name := uuid.New().String()
	path, err := s.metadata.PutExecutorRawConfig(_type, name, config)
	if err != nil {
		return err
	}

	s.rwlock.Lock()
	err = s.addExecutor(_type, path)
	s.rwlock.Unlock()
	if err != nil {
		rerr := s.metadata.RemoveExecutorConfigPath(_type, path)
		if rerr != nil {
			log.Error("Failed to remove executor config path: %s error: %s", path, rerr)
		}
		return err
	}

	return nil
}

func (s *executorService) addExecutor(_type, path string) error {
	exec, err := StartExecutorChildProcess(path, "") // TODO fix me serverAddr
	if err != nil {
		return err
	}
	name := exec.Name()
	_, ok := s.executors[name]
	if ok {
		// exec.Stop() TODO stop it?
		return fmt.Errorf("Executor: %s is already register", name)
	}
	s.executors[name] = Executor{
		Executor:   exec,
		Type:       _type,
		ConfigPath: path,
	}

	return nil
}

func (s *executorService) Remove(names ...string) error {
	s.rwlock.Lock()
	defer s.rwlock.Unlock()

	var eg ErrorGroup
	for _, name := range names {
		exec, ok := s.executors[name]
		if ok {
			err := s.metadata.RemoveExecutorConfigPath(exec.Type, exec.ConfigPath)
			if err != nil {
				eg = append(eg, err)
			}

			exec.Executor.Stop()
			delete(s.executors, exec.Type)
		}
	}

	return eg.Error()
}

func (s *executorService) Find(name string) (*proto.ExecutorView, error) {
	s.rwlock.Lock()
	defer s.rwlock.Unlock()

	exec, ok := s.executors[name]
	if !ok {
		return nil, errors.New("Not found executor " + name)
	}
	return convertExecutor2ExecutorView(exec), nil
}

func (s *executorService) Recreate(name string, config []byte) error {
	//s.rwlock.Lock()
	//defer s.rwlock.Unlock()
	//
	//pipe, ok := s.executors[name]
	//if !ok {
	//	return errors.New("Not found executor " + name)
	//}
	//
	//state := pipe.State()
	//
	//s.removeExecutor(pipe.Name(), pipe.ConfigPath)
	//
	//data, err := yaml.Marshal(conf)
	//if err != nil {
	//	return err
	//}
	//
	//err = s.metadata.Overwrite(proto.FileTypeExecutorConfig, pipe.ConfigPath, data)
	//if err != nil {
	//	return err
	//}
	//
	//err = s.addExecutor(name, pipe.ConfigPath)
	//if err != nil {
	//	return err
	//}
	//
	//if state != executor.Running {
	//	return nil
	//}
	//return pipe.Start()
	return nil
}

func (s *executorService) List() ([]proto.ExecutorView, error) {
	var res []proto.ExecutorView

	s.rwlock.RLock()
	for _, p := range s.executors {
		res = append(res, *convertExecutor2ExecutorView(p))
	}
	s.rwlock.RUnlock()

	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})

	return res, nil
}

func (s *executorService) Control(cmd proto.ControlCommand, names ...string) error {
	s.rwlock.Lock()
	defer s.rwlock.Unlock()

	for _, name := range names {
		exec, ok := s.executors[name]
		if !ok {
			return errors.New("Not found executor " + name)
		}

		switch cmd {
		case proto.ControlCommandStart:
			if err := exec.Start(); err != nil {
				return err
			}
		case proto.ControlCommandStop:
			exec.Stop()
		case proto.ControlCommandRestart:
			exec.Stop()
			if err := exec.Start(); err != nil {
				return err
			}
		default:
			return errors.New("Unsupported method " + string(cmd))
		}
	}

	return nil
}

func (s *executorService) Visualize(format proto.VisualizeFormat, name string) ([]byte, error) {
	//exec, ok := s.executors[name]
	//if !ok {
	//	return nil, errors.New("Not found executor " + name)
	//}
	//
	//buff := bytes.NewBuffer(nil)
	//err := exec.Visualize(buff, string(format))
	//if err != nil {
	//	return nil, err
	//}
	//
	//return buff.Bytes(), nil
	return nil, nil
}

func convertExecutor2ExecutorView(p Executor) *proto.ExecutorView {
	return nil
	//var streamError string
	//var streamErrorCount int
	//p.Monitor().Do(func(root, namespace string, kv expvar.KeyValue) {
	//	switch kv.Key {
	//	case pipeliner.METRICS_KEY_STREAM_ERROR:
	//		streamError = kv.Value.String()
	//	case pipeliner.METRICS_KEY_STREAM_ERROR_COUNT:
	//		i, _ := strconv.Atoi(kv.Value.String())
	//		streamErrorCount += i
	//	}
	//
	//})
	//
	//return &proto.ExecutorView{
	//	Name:          p.Name(),
	//	State:         p.State().String(),
	//	Schedule:      p.Config().Schedule,
	//	Bootstrap:     p.Config().Bootstrap,
	//	StartTime:     p.Monitor().Get(pipeliner.METRICS_KEY_EXECUTOR_START_TIME).String(),
	//	ExitTime:      p.Monitor().Get(pipeliner.METRICS_KEY_EXECUTOR_EXIT_TIME).String(),
	//	RunTimes:      p.Monitor().Get(pipeliner.METRICS_KEY_EXECUTOR_RUN_TIMES).String(),
	//	NextRunTime:   p.Monitor().Get(pipeliner.METRICS_KEY_EXECUTOR_NEXT_RUN_TIME).String(),
	//	LastStartTime: p.Monitor().Get(pipeliner.METRICS_KEY_EXECUTOR_LAST_START_TIME).String(),
	//	LastEndTime:   p.Monitor().Get(pipeliner.METRICS_KEY_EXECUTOR_LAST_END_TIME).String(),
	//	Components:    convertComponents(p.ListComponents()),
	//	Processors:    convertProcessors(p.ListProcessors()),
	//	RawConfig:     mustMarshalConfig(p.GetConfig()),
	//	Error: func() string {
	//		if p.Error() != nil {
	//			return p.Error().Error()
	//		}
	//		return ""
	//	}(),
	//	StreamError:      streamError,
	//	StreamErrorCount: streamErrorCount,
	//}
}

//func mustMarshalConfig(config executor.Config) []byte {
//	b, _ := yaml.Marshal(config)
//	return b
//}

func convertComponents(comps []executor.Component) []proto.ComponentView {
	var res []proto.ComponentView
	for _, c := range comps {
		res = append(res, proto.ComponentView{
			Name:         c.Name,
			RawConfig:    c.RawConfig,
			SampleConfig: c.Factory.SampleConfig(),
			Description:  c.Factory.Description(),
			ReflectType:  fmt.Sprint(c.Factory.ExampleType()),
			InjectName:   c.Component.Instance().Name(),
			ReflectValue: c.Component.Instance().Value().String(),
		})
	}

	return res
}

func convertProcessors(procs []executor.Processor) []proto.ProcessorView {
	var res []proto.ProcessorView
	for _, c := range procs {
		res = append(res, proto.ProcessorView{
			Name:         c.Name,
			RawConfig:    c.RawConfig,
			Description:  c.Factory.Description(),
			SampleConfig: c.Factory.SampleConfig(),
		})
	}
	return res
}
