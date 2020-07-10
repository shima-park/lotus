package component

import (
	"flag"
	"os"
	"sort"
	"strings"

	"github.com/docker/docker/pkg/reexec"
	_ "github.com/shima-park/lotus/pkg/core/component/include"
	"github.com/shima-park/lotus/pkg/rpc/proto"
	cp "github.com/shima-park/lotus/pkg/rpc/service/child_process"
	"github.com/shima-park/seed/component"
	"github.com/shima-park/seed/plugin"
)

const (
	CHILD_PROCESS_LIST_COMPONENTS = "list_components"
)

func init() {
	reexec.Register(CHILD_PROCESS_LIST_COMPONENTS, listComponents)
	if reexec.Init() {
		os.Exit(0)
	}
}

func listComponents() {
	var pluginPaths = flag.String("plugin_paths", "", "List of plugin paths")
	var matches = flag.String("matches", "", "Return a list of components with matching names")

	flag.Parse()

	for _, path := range strings.Split(*pluginPaths, ",") {
		if strings.TrimSpace(path) == "" {
			continue
		}

		_, err := plugin.LoadPlugins(path)
		cp.Failed(err)
	}

	var factoryList []component.NamedFactory
	if *matches != "" {
		for _, match := range strings.Split(*matches, ",") {
			if strings.TrimSpace(match) == "" {
				continue
			}

			f, err := component.GetFactory(match)
			cp.Failed(err)

			factoryList = append(factoryList, component.NamedFactory{
				Name:    match,
				Factory: f,
			})
		}
	} else {
		factoryList = append(factoryList, component.ListFactory()...)
	}

	var results []proto.ComponentView
	for _, f := range factoryList {
		results = append(results, proto.ComponentView{
			Name:         f.Name,
			SampleConfig: f.Factory.SampleConfig(),
			Description:  f.Factory.Description(),
		})
	}

	cp.Success(results)
}

func ListComponents(pluginPaths, matches []string) ([]proto.ComponentView, error) {
	var components []proto.ComponentView
	err := cp.RunCmd(
		CHILD_PROCESS_LIST_COMPONENTS,
		[]string{
			"--plugin_paths", strings.Join(pluginPaths, ","),
			"--matches", strings.Join(matches, ","),
		},
		&components,
	)

	sort.Slice(components, func(i, j int) bool {
		return components[i].Name < components[j].Name
	})

	return components, err
}
