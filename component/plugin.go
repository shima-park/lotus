package component

import (
	"errors"

	"github.com/shima-park/lotus/common/plugin"
)

type componentPlugin struct {
	name    string
	factory Factory
}

const pluginKey = "component"

func init() {
	plugin.MustRegisterLoader(pluginKey, func(ifc interface{}) (string, error) {
		p, ok := ifc.(componentPlugin)
		if !ok {
			return "", errors.New("plugin does not match component plugin type")
		}

		if p.factory != nil {
			if err := Register(p.name, p.factory); err != nil {
				return p.name, err
			}
		}

		return p.name, nil
	})
}

func Plugin(
	module string,
	factory Factory,
) map[string][]interface{} {
	return plugin.MakePlugin(pluginKey, componentPlugin{module, factory})
}
