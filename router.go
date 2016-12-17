package servicerouter

import (
	"errors"
	"golang.org/x/net/context"
	"regexp"
	"strings"
	"sync"
)

type routeMatcher interface {
	match(string) (string, bool)
}

// ErrRouteNotFound route was not found
var ErrRouteNotFound = errors.New("No Route Found")

// Route holds a match
type Route struct {
	mu      sync.Mutex
	name    string
	matcher routeMatcher
	sub     []*Route
	h       RouteHandler
}

// RoutedContext holds context
type RoutedContext struct {
	context.Context
	Route *Route
	Path  string
}

// RouteHandler callback on matched route
type RouteHandler interface {
	Handle(*RoutedContext) (interface{}, error)
}

// Router : Service Router
type Router struct {
	rw     sync.RWMutex
	routes []*Route
}

// RouteHandlerFunc simplified Handler interface
type RouteHandlerFunc func(*RoutedContext) (interface{}, error)

// Handle implementation of RoutedHandler
func (f RouteHandlerFunc) Handle(rc *RoutedContext) (interface{}, error) {
	r, e := f(rc)
	return r, e
}

// NewRouter Create new router
func NewRouter() *Router {
	return &Router{
		routes: make([]*Route, 0, 1),
	}
}

// ExecPath route the specified path
func (r *Router) ExecPath(path string) (interface{}, error) {
	return r.Exec(path, context.Background())
}

// Exec execute routing
func (r *Router) Exec(path string, ctx context.Context) (interface{}, error) {
	r.rw.RLock()
	defer r.rw.RUnlock()

	for _, route := range r.routes {
		matched, isrouted := route.matchRoute(path)
		if isrouted {
			rctx := &RoutedContext{
				Context: ctx,
				Path:    path,
				Route:   matched,
			}

			rz, err := matched.h.Handle(rctx)
			return rz, err
		}
	}

	return nil, ErrRouteNotFound
}

// Clear remove all configured routes
func (r *Router) Clear() *Router {
	r.routes = make([]*Route, 0, 1)
	return r
}

// SimpleRoute litral string matching ie: home.matching == home.matching
func (r *Router) SimpleRoute(path string) *Route {
	return r.newRoute().matchSimple(path)
}

// PrefixRoute string prefix matching home. == home.matching
func (r *Router) PrefixRoute(prefix string) *Route {
	return r.newRoute().matchPrefix(prefix)
}

// RegExpRoute regular expression based matching
func (r *Router) RegExpRoute(re *regexp.Regexp) *Route {
	return r.newRoute().matchRegExp(re)
}

func (r *Router) newRoute() *Route {
	r.rw.Lock()
	defer r.rw.Unlock()

	nroute := &Route{
		sub: make([]*Route, 0),
	}

	r.routes = append(r.routes, nroute)
	return nroute
}

// Name name of route. just for logging and debuging
func (r *Route) Name() string {
	return r.name
}

// SetName name of route. just for logging and debuging
func (r *Route) SetName(name string) *Route {
	r.name = name
	return r
}

// Handler callback if route matches
func (r *Route) Handler(f RouteHandler) *Route {
	r.h = f
	return r
}

// HandlerFunc callback if route matches
func (r *Route) HandlerFunc(f RouteHandlerFunc) *Route {
	return r.Handler(f)
}

// SimpleSubRoute litral string matching ie: home.matching == home.matching
func (r *Route) SimpleSubRoute(path string) *Route {
	return r.subRoute().matchSimple(path)
}

// PrefixSubRoute string prefix matching home. == home.matching
func (r *Route) PrefixSubRoute(prefix string) *Route {
	return r.subRoute().matchPrefix(prefix)
}

// RegExpSubRoute regular expression based matching
func (r *Route) RegExpSubRoute(re *regexp.Regexp) *Route {
	return r.subRoute().matchRegExp(re)
}

func (r *Route) matchPrefix(prefix string) *Route {
	r.matcher = simplePrefixMatcher(prefix)
	return r
}

func (r *Route) matchSimple(path string) *Route {
	r.matcher = simpleMatcher(path)
	return r
}

func (r *Route) matchRegExp(rx *regexp.Regexp) *Route {
	r.matcher = regexpMatcher(rx.Copy())
	return r
}

func (r *Route) subRoute() *Route {
	r.mu.Lock()
	defer r.mu.Unlock()

	s := &Route{
		sub: make([]*Route, 0),
	}

	r.sub = append(r.sub, s)
	return s
}

func (r *Route) matchRoute(path string) (*Route, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.matcher == nil || path == "" {
		return nil, false // this route would never match anything
	}

	if extra, matched := r.matcher.match(path); matched {
		if r.sub != nil && extra != "" {
			for _, sroute := range r.sub {
				if nroute, ok := sroute.matchRoute(extra); ok {
					return nroute, true
				}
			}
		}

		if r.h != nil {
			return r, true
		}
	}
	return nil, false
}

type routeMatcherFunc func(string) (string, bool)

func (f routeMatcherFunc) match(p string) (string, bool) {
	r, e := f(p)
	return r, e
}

func simplePrefixMatcher(prefix string) routeMatcherFunc {
	return func(path string) (string, bool) {
		nprefix := strings.TrimPrefix(path, prefix)
		println(nprefix, "=", prefix, path)
		return nprefix, path != nprefix
	}
}

func simpleMatcher(patten string) routeMatcherFunc {
	return func(path string) (string, bool) {
		println(patten, "=", path)
		return "", patten == path
	}
}

func regexpMatcher(re *regexp.Regexp) routeMatcherFunc {
	return func(path string) (string, bool) {
		if idx := re.FindStringIndex(path); idx != nil {
			return path[idx[1]:], true
		}

		return "", false
	}
}
