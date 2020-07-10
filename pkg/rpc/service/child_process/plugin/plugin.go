package plugin

import (
	"flag"
	"os"
	"sort"
	"strings"

	"github.com/docker/docker/pkg/reexec"
	"github.com/shima-park/lotus/pkg/rpc/proto"
	cp "github.com/shima-park/lotus/pkg/rpc/service/child_process"
	"github.com/shima-park/seed/plugin"
)

const (
	CHILD_PROCESS_LIST_PLUGINS = "list_plugins"
)

func init() {
	reexec.Register(CHILD_PROCESS_LIST_PLUGINS, listPlugins)
	if reexec.Init() {
		os.Exit(0)
	}
}

func listPlugins() {
	var paths = flag.String("paths", "", "List of plugin paths")
	var matches = flag.String("matches", "", "Return a list of plugins with matching names")

	flag.Parse()

	plugins := map[string]proto.PluginView{}
	for _, path := range strings.Split(*paths, ",") {
		if strings.TrimSpace(path) == "" {
			continue
		}

		ps, err := plugin.LoadPlugins(path)
		cp.Failed(err)

		for _, p := range ps {
			plugins[p.Name] = proto.PluginView{
				Name:     p.Name,
				Path:     p.Path,
				Module:   p.Module,
				OpenTime: p.OpenTime,
				Error: func() string {
					if p.Error != nil {
						return p.Error.Error()
					}
					return ""
				}(),
			}
		}
	}

	var results []proto.PluginView

	if *matches != "" {
		for _, match := range strings.Split(*matches, ",") {
			if strings.TrimSpace(match) == "" {
				continue
			}

			plug, ok := plugins[match]
			if !ok {
				continue
			}

			results = append(results, plug)
		}
	} else {
		for _, plug := range plugins {
			results = append(results, plug)
		}
	}

	cp.Success(results)
}

func ListPlugins(paths, matches []string) ([]proto.PluginView, error) {
	var plugins []proto.PluginView
	err := cp.RunCmd(
		CHILD_PROCESS_LIST_PLUGINS,
		[]string{
			"--paths", strings.Join(paths, ","),
			"--matches", strings.Join(matches, ","),
		},
		&plugins,
	)

	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})

	return plugins, err
}
