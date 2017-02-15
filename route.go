package servicerouter

import (
	"golang.org/x/net/context"
	"regexp"
	"strings"
	"sync"
)

type routeMatcherFunc func(string) (string, bool)

func (f routeMatcherFunc) match(p string) (string, bool) {
	r, e := f(p)
	return r, e
}

func simplePrefixMatcher(prefix string) routeMatcherFunc {
	return func(path string) (string, bool) {
		nprefix := strings.TrimPrefix(path, prefix)
		return nprefix, path != nprefix
	}
}

func simpleMatcher(patten string) routeMatcherFunc {
	return func(path string) (string, bool) {
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

// Handler callback on matched route
type Handler interface {
	Handle(context.Context, interface{}) (interface{}, error)
}

// HandlerFunc simplified Handler interface
type HandlerFunc func(context.Context, interface{}) (interface{}, error)

// Handle implementation of RoutedHandler
func (f HandlerFunc) Handle(ctx context.Context, req interface{}) (r interface{}, e error) {
	r, e = f(ctx, req)
	return
}

// Route holds a match
type Route struct {
	mu      sync.Mutex
	name    string
	matcher routeMatcher
	sub     []*Route
	h       Handler
}

// Name name of route. just for logging and debuging
func (r *Route) Name() string {
	return r.name
}

// AddRoute add a branch to this routing three
func (r *Route) AddRoute(opts ...RouteOptFn) *Route {
	r.mu.Lock()
	defer r.mu.Unlock()

	s := newRoute(opts...)
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

type RouteOptFn func(*Route)

func newRoute(opts ...RouteOptFn) *Route {
	s := &Route{
		name: "UnNamed Route",
		sub:  make([]*Route, 0),
	}

	for _, opt := range opts {
		opt(s)
	}
	return s
}

// SimpleRoute litral string matching ie: home.matching == home.matching
func SimpleRoute(path string) RouteOptFn {
	return func(r *Route) {
		r.matcher = simpleMatcher(path)
	}
}

// PrefixRoute string prefix matching home. == home.matching
func PrefixRoute(prefix string) RouteOptFn {
	return func(r *Route) {
		r.matcher = simplePrefixMatcher(prefix)
	}
}

// RegExpRoute regular expression based matching
func RegExpRoute(re *regexp.Regexp) RouteOptFn {
	return func(r *Route) {
		r.matcher = regexpMatcher(re)
	}
}

// Name set name. for debuging purposes only
func Name(name string) RouteOptFn {
	return func(r *Route) {
		r.name = name
	}
}

// RouteHandler set callback to execute if route matches
func RouteHandler(f Handler) RouteOptFn {
	return func(r *Route) {
		r.h = f
	}
}

// RouteHandlerFunc set callback to execute if route matches
func RouteHandlerFunc(f HandlerFunc) RouteOptFn {
	return func(r *Route) {
		r.h = f
	}
}
