package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	// This is `httprouter`. Ensure to install it first via `go get`.
	"github.com/julienschmidt/httprouter"
)

// We need a data store. For our purposes, a simple map
// from string to string is completely sufficient.
type store struct {
	data map[string]string

	// Handlers run concurrently, and maps are not thread-safe.
	// This mutex is used to ensure that only one goroutine can update `data`.
	m sync.RWMutex
}

var (
	// We need a flag for setting the listening address.
	// We set the default to port 8080, which is a common HTTP port
	// for servers with local-only access.
	addr = flag.String("addr", ":8080", "http service address")

	// Now we create the data store.
	s = store{
		data: map[string]string{},
		m:    sync.RWMutex{},
	}
)

// ## main
func main() {
	// The main function starts by parsing the commandline.
	flag.Parse()

	// Now we can create a new `httprouter` instance...
	r := httprouter.New()

	// ...and add some routes.
	// `httprouter` provides functions named after HTTP verbs.
	// So to create a route for HTTP GET, we simply need to call the `GET` function
	// and pass a route and a handler function.
	// The first route is `/entry` followed by a key variable denoted by a leading colon.
	// The handler function is set to `show`.
	r.GET("/entry/:key", show)

	// We do the same for `/list`. Note that we use the same handler function here;
	// we'll switch functionality within the `show` function based on the existence
	// of a key variable.
	r.GET("/list", show)

	// For updating, we need a PUT operation. We want to pass a key and a value to the URL,
	// so we add two variables to the path. The handler function for this PUT operation
	// is `update`.
	r.PUT("/entry/:key/:value", update)

	// Finally, we just have to start the http Server. We pass the listening address
	// as well as our router instance.
	err := http.ListenAndServe(*addr, r)

	// For this demo, let's keep error handling simple.
	// `log.Fatal` prints out an error message and exits the process.
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

// ## The handler functions

// Let's implement the show function now. Typically, handler functions receive two parameters:
//
// * A Response Writer, and
// * a Request object.
//
// `httprouter` handlers receive a third parameter of type `Params`.
// This way, the handler function can access the key and value variables
// that have been extracted from the incoming URL.
func show(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	// To access these parameters, we call the `ByName` method, passing the variable name that we chose when defining the route in `main`.
	k := p.ByName("key")

	// The show function serves two purposes.
	// If there is no key in the URL, it lists all entries of the data map.
	if k == "" {
		// Lock the store for reading.
		s.m.RLock()
		fmt.Fprintf(w, "Read list: %v", s.data)
		s.m.RUnlock()
		return
	}

	// If a key is given, the show function returns the corresponding value.
	// It does so by simply printing to the ResponseWriter parameter, which
	// is sufficient for our purposes.
	s.m.RLock()
	fmt.Fprintf(w, "Read entry: s.data[%s] = %s", k, s.data[k])
	s.m.RUnlock()
}

// The update function has the same signature as the show function.
func update(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	// Fetch key and value from the URL parameters.
	k := p.ByName("key")
	v := p.ByName("value")

	// We just need to either add or update the entry in the data map.
	s.m.Lock()
	s.data[k] = v
	s.m.Unlock()

	// Finally, we print the result to the ResponseWriter.
	fmt.Fprintf(w, "Updated: s.data[%s] = %s", k, v)
}
