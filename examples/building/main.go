package main

import (
	"fmt"
	sr "github.com/codmajik/servicerouter"
	"net/http"
)

func handleAnyOtherRequest(ctx *sr.RoutedContext) (interface{}, error) {
	return fmt.Sprintf("don't think you want '%s'", ctx.Path), nil
}
func handleTwoStory(ctx *sr.RoutedContext) (interface{}, error) {
	return "TWO_STORY_BUILDING", nil
}

func handleOneStory(ctx *sr.RoutedContext) (interface{}, error) {
	return "ONE_STORY_BUILDING", nil
}

func handleNoBuilding(ctx *sr.RoutedContext) (interface{}, error) {
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
