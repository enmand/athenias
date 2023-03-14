package athenais

import "maunium.net/go/mautrix/event"

type RouteHandler func(*event.Event) error

// Route defines a plugin route handler, based on event type and room ID
type Route struct {
	// EventType is the event type to match
	EventType event.Type

	// Handler is the handler to call
	Handler RouteHandler
}

func NewRoute(eventType event.Type, handler RouteHandler) Route {
	return Route{
		EventType: eventType,
		Handler:   handler,
	}
}

// Router is a router for plugin routes
type Router struct {
	routes          []Route
	routeEventCache map[event.Type][]Route
}

func NewRouter() *Router {
	return &Router{
		routes:          make([]Route, 0),
		routeEventCache: make(map[event.Type][]Route),
	}
}

func (r *Router) AddRoute(route Route) {
	r.routes = append(r.routes, route)
	r.routeEventCache[route.EventType] = append(r.routeEventCache[route.EventType], route)
}

func (r *Router) GetRoutesByEvent(eventType event.Type) []Route {
	return r.routeEventCache[eventType]
}

func (r *Router) GetRoutes() []Route {
	return r.routes
}

func (r *Router) Handle(evt *event.Event) error {
	for _, route := range r.GetRoutesByEvent(evt.Type) {
		if err := route.Handler(evt); err != nil {
			return err
		}
	}

	return nil
}
