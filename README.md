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
  "net/http"
  "strings"
  "github.com/codmajik/servicerouter"
  
)


function main() {

  // NOTE: The routing scheme used here is arbitrary
  router := servicerouter.NewRouter()
  
  router.SimpleRoute("home").
    HandlerFunc(func(ctx *servicerouter.RoutedContext) (interface{}, error) {
      return "HOME_CONTENT", nil
    })
    
  // use prefix to do subrouting
  subroute := router.PrefixRoute("users.").
    HandlerFunc(func(ctx *servicerouter.RoutedContext) (interface{}, error) {
      return "USERS_INVALID_OPERATION", nil
    })
    
    
  // the following routes are sub of users, so  users. would have to match first
  
  // users.list --> fetch all users
  subroute.SimpleSubRoute("list").
    HandlerFunc(func(ctx *servicerouter.RoutedContext) (interface{}, error) {
      return "LIST_OF_USERS", nil
    })
  
  //users.list //
   subroute.PrefixSubRoute("list.groups/").
      HandlerFunc(func(ctx *servicerouter.RoutedContext) (interface{}, error) {
        return "LIST_OF_USERS_BY_GROUP", nil
      }).
      SimpleSubRoute("admin"). // you can even subroute on a subroute - META ZEN
        HandlerFunc(func(ctx *servicerouter.RoutedContext) (interface{}, error) {
          return "LIST_OF_ADMIN_USERS", nil
        }).
  
  // users.save
  subroute.SimpleSubRoute("save").
      HandlerFunc(func(ctx *servicerouter.RoutedContext) (interface{}, error) {
        params := ctx.Context.Value("params").(map[string]string)
        fmt.Println("save to db", params)
        return "HOUSTON_ALL_IS_WELL", nil
      })
      
      
   
   
   http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
      qpath := req.URL.Query().Get("action")
      if (qpath == "") {
        qpath = strings.Replace(req.RequestURI, "/", ".", 1)
      }
      
      // if you have parameters to pass call with a
      // ctx := context.WithValue(req.Context(), SOME_KEY, params) 
      //result, err := router.Exec(qpath, ctx)
      result, err := router.ExecPath(qpath)
      if err != nil {
        rw.Write([]byte("Path Not Found: " + err.Error()))
        return
      }

      rw.Write([]byte(result.(string)))
    })

	fmt.Println("Listening for http request on :8080")
	fmt.Println(http.ListenAndServe(":8080", nil))
}
```

#### Todo
- [ ] More documentation
- [ ] More examples
- [ ] Maybe a tutorial



#### License
MIT  -- See LICENSE file

