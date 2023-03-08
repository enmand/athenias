package athenais

import (
	"github.com/rs/zerolog"
)

var plugins map[string]Plugin

type Plugin interface {
	// Name returns the name of the plugin
	Name() string

	// Init initializes the plugin
	Init(*Bot, *zerolog.Logger)
}

func Register(plug Plugin) {
	name := plug.Name()
	if _, ok := plugins[name]; ok {
		panic("plugin already registered: " + name)
	}

	plugins[name] = plug
}
