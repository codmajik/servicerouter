package servicerouter

import (
	"regexp"
	"testing"

	"golang.org/x/net/context"
)

type testCase struct {
	path   string
	result interface{}
	err    error
}

func (r *testCase) test(t *testing.T, router *Router) {
	rvalue, err := router.ExecPath(r.path, nil)

	if r.err != err {
		t.Error("expected error", r.err, "got", err)
	}

	if rvalue != r.result {
		t.Error("expected result", r.result, "got", rvalue)
	}
}

const (
	MatchTypeSimple = "SIMPLE_MATCH"
	MatchTypePrefix = "PREFIX_MATCH"
	MatchTypeRegex  = "REGEX_MATCH"
)

type routeTestCase struct {
	matchType  string
	route      string
	tcase      *testCase
	rootPrefix string
}

func (r *routeTestCase) hFunc(ctx context.Context, req interface{}) (interface{}, error) {
	return r.tcase.result, r.tcase.err
}

func (r *routeTestCase) test(t *testing.T) {
	router := NewRouter(RootPrefix(r.rootPrefix))

	switch r.matchType {
	case MatchTypeSimple:
		router.SimpleRoute(r.route).HandlerFunc(r.hFunc)
	case MatchTypePrefix:
		router.PrefixRoute(r.route).HandlerFunc(r.hFunc)
	case MatchTypeRegex:
		re, err := regexp.Compile(r.route)
		if err != nil {
			t.Fatal("Bad Test Case -- RegExp compile failed", err)
			t.FailNow()
			return
		}

		router.RegExpRoute(re).HandlerFunc(r.hFunc)
	default:
		t.Fatal("invalid test case --- unsupported match type ", r.matchType)
		t.FailNow()
		return
	}

	r.tcase.test(t, router)
}

// func runTest(testcases []*routeTestCase, t *testing.T, router *Router) {

// 	for _, r := range testcases {
// 		r.test(t, router)
// 	}

// }

func TestSimpleRoute(t *testing.T) {
	t.Parallel()

	testcases := []*routeTestCase{
		{
			matchType: MatchTypeSimple,
			route:     "simpleroute",
			tcase: &testCase{
				path:   "simpleroute",
				err:    nil,
				result: "This is value",
			},
		},
		{
			matchType: MatchTypeSimple,
			route:     "simpleroute",
			tcase: &testCase{
				path:   "badpath",
				err:    ErrRouteNotFound,
				result: nil,
			},
		},
	}

	for _, c := range testcases {
		c.test(t)
	}
}

func TestPrefixRoute(t *testing.T) {
	t.Parallel()

	testcases := []*routeTestCase{
		{
			matchType: MatchTypePrefix,
			route:     "service.",
			tcase: &testCase{
				path:   "service.profiles",
				err:    nil,
				result: "User Profile",
			},
		},
		{
			matchType: MatchTypePrefix,
			route:     "service.",
			tcase: &testCase{
				path:   "service",
				err:    ErrRouteNotFound,
				result: nil,
			},
		},
	}

	for _, c := range testcases {
		c.test(t)
	}
}

func TestRegExpRoute(t *testing.T) {
	t.Parallel()

	testcases := []*routeTestCase{
		{
			matchType: MatchTypeRegex,
			route:     "service[.](profiles|users)[.]?.*",
			tcase: &testCase{
				path:   "service.profiles",
				err:    nil,
				result: "User Profile",
			},
		},
		{
			matchType: MatchTypeRegex,
			route:     "service[.](profiles|users)[.]?.*",
			tcase: &testCase{
				path:   "service.users",
				err:    nil,
				result: "User Profile",
			},
		},
		{
			matchType: MatchTypeRegex,
			route:     "service[.](profiles|users)[.]?.*",
			tcase: &testCase{
				path:   "service.logs",
				err:    ErrRouteNotFound,
				result: nil,
			},
		},
	}

	for _, c := range testcases {
		c.test(t)
	}
}

func TestRouterRootPrefix(t *testing.T) {
	t.Parallel()
	testcases := []*routeTestCase{
		{
			matchType:  MatchTypePrefix,
			route:      "service.",
			rootPrefix: "myrootpath.",
			tcase: &testCase{
				path:   "myrootpath.service.profiles",
				err:    nil,
				result: "User Profile",
			},
		},
		{
			matchType:  MatchTypePrefix,
			route:      "service.",
			rootPrefix: "myrootpath.",
			tcase: &testCase{
				path:   "myroot.service.profiles",
				err:    ErrRouteNotFound,
				result: nil,
			},
		},
		{
			matchType:  MatchTypePrefix,
			route:      "service.",
			rootPrefix: "myrootpath.",
			tcase: &testCase{
				path:   "myrootpath.service",
				err:    ErrRouteNotFound,
				result: nil,
			},
		},
	}

	for _, c := range testcases {
		c.test(t)
	}
}

func TestNestedRoutes(t *testing.T) {
	t.Parallel()
}
