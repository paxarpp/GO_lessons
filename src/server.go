package main

import (
	"fmt"
	"net/http"
	"strings"
	"log"
)

func homeRouterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParceForm()
	fmt.Println(r.Form)
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "hello, what are you?")

}

func main() {
	http.HendleFunc("/", homeRouterHandler)
	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}