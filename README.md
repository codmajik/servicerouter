# Service Router
A simple transport and schema agnostic router for your services. 

#### Introduction
[TODO] Write This

#### Why
Am sure like me, you have tried creating service that interact over [WebSocket](https://wikipedia.org/wiki/WebSocket) or other transports that don't have idea of routes (eg: Plain TCP).

You probably have also built a service on http and want to switch to [RabbitMQ](https://www.rabbitmq.com/), and now you have to rewrite everything; because you have to subscribe 500 topics for each feature or come up with a scheme to call appropriate methods.

With servicerouter you implement your own scheme which would be portable across transports.

Now you can send
```json
{
  "action":"profile.create",
  "parameters": {
    "firstName": "John",
    "lastName": "Doe",
    "location": "Accra",
    "country": "Ghana"
  }
 }
 ```

or 

`
  http://service?action=profile.create&firstName=John&lastName=Doe&location=Accra&country=Ghana
`

And handle all by changing or adding a few lines of code

#### Getting started

```shell
go get github.com/codmajik/servicerouter
```

##### sample app
```golang
package main

import (
	"fmt"
	sr "github.com/codmajik/servicerouter"
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

func main() {

	// NOTE: The routing scheme used here is arbitrary
	router := sr.NewRouter()

	router.AddRoute( 
		sr.SimpleRoute("home"),
		sr.RouteHandlerFunc(func(ctx context.Context, req interface{}) (interface{}, error) {
			return "HOME_CONTENT", nil
		}),
	)

	// use prefix to do subrouting
	subroute := router.AddRoute(
		sr.PrefixRoute("users."),
		sr.RouteHandlerFunc(func(ctx context.Context, req interface{}) (interface{}, error) {
			return "USERS_INVALID_OPERATION", nil
		}),
	)

	// the following routes are sub of users, so  users. would have to match first

	// users.list --> fetch all users
	subroute.AddRoute(
		sr.SimpleRoute("list"),
		sr.RouteHandlerFunc(func(ctx context.Context, req interface{}) (interface{}, error) {
			return "LIST_OF_USERS", nil
		})
	)

		//users.list //
	subroute.AddRoute(
		sr.Name("List Users by Group Association")
		sr.PrefixRoute("list.groups/"),
		sr.RouteHandlerFunc(func(ctx context.Context, req interface{}) (interface{}, error) {
			return "LIST_OF_USERS_BY_GROUP", nil
		}),
	).AddRoute(
		sr.SimpleRoute("admin"), // you can even subroute on a subroute - META ZEN
		sr.RouteHandlerFunc(func(ctx context.Context, req interface{}) (interface{}, error) {
			return "LIST_OF_ADMIN_USERS", nil
		}),
	)

	// users.save
	subroute.AddRoute( 
		sr.SimpleRoute("save"),
		sr.RouteHandlerFunc(func(ctx context.Context, req interface{}) (interface{}, error) {
			params := req.(map[string]string)
			fmt.Println("save to db", params)
			return "HOUSTON_ALL_IS_WELL", nil
		}),
	)

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {

		qpath = strings.Replace(req.RequestURI, "/", ".", -1)

		// result, err := router.Exec(req.Context(), qpath, nil)
		result, err := router.ExecPath(qpath, nil)
		if err != nil {
			rw.Write([]byte("Path Not Found: " + err.Error()))
			return
		}

		if result != nil {
			rw.Write([]byte(result.(string)))
		}
	})

	fmt.Println("Listening for http request on :8080")
	fmt.Println(http.ListenAndServe(":8080", nil))
}


```
#### Notes
* SimpleRoute: don't add sub routes to simple routes, sub routes are only matched after the base match
* RouteHandlerFunc: match gokit's endpoint.Endpoint and can be composed


#### Inspired by
- [GoKit](http://gokit.io)
- [Gorrila Mux](http://www.gorillatoolkit.org/pkg/mux)


#### Todo
- [ ] More documentation
- [ ] More examples
- [ ] Maybe a tutorial


#### License
MIT  -- See LICENSE file

