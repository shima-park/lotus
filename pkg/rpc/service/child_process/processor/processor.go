package processor

import (
	"flag"
	"os"
	"sort"
	"strings"

	"github.com/docker/docker/pkg/reexec"
	"github.com/shima-park/lotus/pkg/rpc/proto"
	cp "github.com/shima-park/lotus/pkg/rpc/service/child_process"
	"github.com/shima-park/seed/plugin"
	"github.com/shima-park/seed/processor"
)

const (
	CHILD_PROCESS_LIST_PROCESSORS = "list_processors"
)

func init() {
	reexec.Register(CHILD_PROCESS_LIST_PROCESSORS, listProcessors)
	if reexec.Init() {
		os.Exit(0)
	}
}

func listProcessors() {
	var pluginPaths = flag.String("plugin_paths", "", "List of plugin paths")
	var matches = flag.String("matches", "", "Return a list of processors with matching names")

	flag.Parse()

	for _, path := range strings.Split(*pluginPaths, ",") {
		if strings.TrimSpace(path) == "" {
			continue
		}

		_, err := plugin.LoadPlugins(path)
		cp.Failed(err)
	}

	var factoryList []processor.NamedFactory
	if *matches != "" {
		for _, match := range strings.Split(*matches, ",") {
			if strings.TrimSpace(match) == "" {
				continue
			}

			f, err := processor.GetFactory(match)
			cp.Failed(err)

			factoryList = append(factoryList, processor.NamedFactory{
				Name:    match,
				Factory: f,
			})
		}
	} else {
		factoryList = append(factoryList, processor.ListFactory()...)
	}

	var results []proto.ProcessorView
	for _, f := range factoryList {
		results = append(results, proto.ProcessorView{
			Name:         f.Name,
			SampleConfig: f.Factory.SampleConfig(),
			Description:  f.Factory.Description(),
		})
	}

	cp.Success(results)
}

func ListProcessors(pluginPaths, matches []string) ([]proto.ProcessorView, error) {
	var processors []proto.ProcessorView
	err := cp.RunCmd(
		CHILD_PROCESS_LIST_PROCESSORS,
		[]string{
			"--plugin_paths", strings.Join(pluginPaths, ","),
			"--matches", strings.Join(matches, ","),
		},
		&processors,
	)

	sort.Slice(processors, func(i, j int) bool {
		return processors[i].Name < processors[j].Name
	})

	return processors, err
}
