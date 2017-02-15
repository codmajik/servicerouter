package servicerouter

import (
	"errors"
	"golang.org/x/net/context"
	"strings"
	"sync"
)

const (
	//RouteKey Context value key for provided path
	RouteKey = "sr.Route:Key"

	//RouteName Context value key for provided path
	RouteName = "sr.Route:Name"
)

type routeMatcher interface {
	match(string) (string, bool)
}

// ErrRouteNotFound route was not found
var ErrRouteNotFound = errors.New("Route Not Found")

// OptFn option function
type OptFn func(*Router)

// Router : Service Router
type Router struct {
	rw        sync.RWMutex
	rootPath  string
	routes    []*Route
	cbRouteFn func(string, *Route)
}

// RootPrefix set rootPrefix for this router
// this is used to add a universal route that is not part of routing scheme
// mount the same routing on different roots without changing your scheme
func RootPrefix(rootPrefix string) OptFn {
	return func(r *Router) {
		r.rootPath = rootPrefix
	}
}

// RouteCallback called when a route matches with info
func RouteCallback(cb func(string, *Route)) OptFn {
	return func(r *Router) {
		r.cbRouteFn = cb
	}
}

// NewRouter Create new router
func NewRouter(opts ...OptFn) *Router {
	r := &Router{
		rootPath: "",
		routes:   make([]*Route, 0, 1),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// RootPrefix get root prefix configured for this router
func (r *Router) RootPrefix() string {
	return r.rootPath
}

// AddRoute add a new route
func (r *Router) AddRoute(opts ...RouteOptFn) *Route {
	r.rw.Lock()
	defer r.rw.Unlock()

	n := newRoute(opts...)
	r.routes = append(r.routes, n)
	return n
}

// ExecPath route the specified path
func (r *Router) ExecPath(path string, req interface{}) (interface{}, error) {
	return r.Exec(context.Background(), path, req)
}

// Exec execute routing
func (r *Router) Exec(ctx context.Context, path string, req interface{}) (interface{}, error) {
	r.rw.RLock()
	defer r.rw.RUnlock()

	if strings.HasPrefix(path, r.rootPath) {
		strippedPath := strings.TrimPrefix(path, r.rootPath)
		for _, route := range r.routes {
			matched, isrouted := route.matchRoute(strippedPath)
			if isrouted {
				if ctx == nil {
					ctx = context.Background()
				}

				ctx = context.WithValue(ctx, RouteName, matched.Name())
				ctx = context.WithValue(ctx, RouteKey, strippedPath)

				if r.cbRouteFn != nil {
					r.cbRouteFn(path, matched)
				}
				rz, err := matched.h.Handle(ctx, req)
				return rz, err
			}
		}
	}

	return nil, ErrRouteNotFound
}

// Clear remove all configured routes
func (r *Router) Clear() *Router {
	r.routes = make([]*Route, 0, 1)
	return r
}
