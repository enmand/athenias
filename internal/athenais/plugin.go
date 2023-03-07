package athenais

type Plugin interface {
	// Name returns the name of the plugin
	Name() string

	// Init initializes the plugin
	Init(*Bot) error
}
