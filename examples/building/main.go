package main

import (
	"fmt"
	sr "github.com/codmajik/servicerouter"
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

func handleAnyOtherRequest(ctx context.Context, req interface{}) (interface{}, error) {
	qvals := ""

	r, ok := req.(*http.Request)
	if ok {
		qvals = r.URL.RawQuery
	}
	return fmt.Sprintf("you want '%s'? => with params %s", ctx.Value(sr.PathKey), qvals), nil
}

func handleTwoStory(ctx context.Context, req interface{}) (interface{}, error) {
	return "TWO_STORY_BUILDING", nil
}

func handleOneStory(ctx context.Context, req interface{}) (interface{}, error) {
	return "ONE_STORY_BUILDING", nil
}

func handleNoBuilding(ctx context.Context, req interface{}) (interface{}, error) {
	return "LIST_ALL_BUILDING", nil
}

func main() {
	router := sr.NewRouter()

	router.SimpleRoute("building").HandlerFunc(handleNoBuilding)
	router.SimpleRoute("building.onestory").HandlerFunc(handleOneStory)

	route := router.PrefixRoute("building.")
	route.PrefixSubRoute("onestory.").HandlerFunc(handleAnyOtherRequest)
	route.SimpleSubRoute("twostory").HandlerFunc(handleTwoStory)

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {

		qpath := req.URL.Query().Get("q")
		if qpath == "" {
			qpath = strings.Trim(req.URL.Path, "/?")
			qpath = strings.Replace(qpath, "/", ".", -1)
		}

		fmt.Println("[REQUEST]", qpath)
		result, err := router.ExecPath(qpath, req)
		if err != nil {
			rw.Write([]byte("Path Not Found: " + err.Error()))
			return
		}

		rw.Write([]byte(result.(string)))
	})

	fmt.Println("Listening for http request on :9091")
	fmt.Println(http.ListenAndServe(":9091", nil))
}
