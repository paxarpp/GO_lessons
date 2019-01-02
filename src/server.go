package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
)

func main() {
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt)

	handler := http.NewServeMux()

	handler.HandleFunc("/book/", BasicAuth(Logger(HandleBook)))

	handler.HandleFunc("/books/", BasicAuth(Logger(HandleBooks)))

	s := &http.Server{
		Addr:           ":8080",
		Handler:        handler,          // if nil use default http.DefaultServeMux
		ReadTimeout:    10 * time.Second, // max duration reading entire request
		WriteTimeout:   10 * time.Second, // max timing write response
		IdleTimeout:    15 * time.Second, // max time wait for the next request
		MaxHeaderBytes: 1 << 20,          // 2^20 or 128kbytes
	}

	log.Printf("Listening on http://%s\n", s.Addr)
	log.Fatal(s.ListenAndServe())

}

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Printf("server [net/http] method [%s]  connection from [%v]", r.Method, r.RemoteAddr)

		next.ServeHTTP(w, r)
	}
}

func BasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 || !validate(pair[0], pair[1]) {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func validate(username, password string) bool {
	if username == "test" && password == "test" { //Basic dGVzdDp0ZXN0
		return true
	}
	return false
}

func HandleBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		HandleGetBooks(w, r)
	}
}

func HandleGetBooks(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	books, _ := json.Marshal(bookStore.GetBooks())

	w.Write(books)
}

func HandleBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		HandleGetBook(w, r)

	} else if r.Method == http.MethodPost {
		HandleAddBook(w, r)

	} else if r.Method == http.MethodPut {
		HandleUpdateBook(w, r)

	} else if r.Method == http.MethodDelete {
		HandleDeleteBook(w, r)

	} else {
		HandleMethodIsNotAllowed(w, r)

	}
}

func HandleMethodIsNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	msg, _ := json.Marshal(fmt.Sprintf("Method %s not allowed", r.Method))
	w.Write(msg)
}

func HandleGetBook(w http.ResponseWriter, r *http.Request) {
	bookid := strings.Replace(r.URL.Path, "/book/", "", 1)

	book := bookStore.FindBookById(bookid)

	if book == nil {
		w.WriteHeader(http.StatusNotFound)
		error, _ := json.Marshal(fmt.Sprintf("Book with id %s not found", bookid))

		w.Write(error)

		return
	}

	w.WriteHeader(http.StatusOK)

	bookidJson, _ := json.Marshal(book)

	w.Write(bookidJson)
}

func HandleAddBook(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var book Book

	err := decoder.Decode(&book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		error, _ := json.Marshal(fmt.Sprintf("Bad request. %v", err))

		w.Write(error)
		return
	}

	err = bookStore.AddBook(book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		error, _ := json.Marshal(fmt.Sprintf("Bad request. %v", err))

		w.Write(error)
		return
	}

	HandleGetBooks(w, r)
}

func HandleUpdateBook(w http.ResponseWriter, r *http.Request) {
	bookid := strings.Replace(r.URL.Path, "/book/", "", 1)

	decoder := json.NewDecoder(r.Body)

	var book Book

	err := decoder.Decode(&book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		error, _ := json.Marshal(fmt.Sprintf("Bad request. %v", err))

		w.Write(error)
		return
	}

	book.Id = bookid

	err = bookStore.SetBook(book)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		error, _ := json.Marshal(fmt.Sprintf("%v", err))

		w.Write(error)

		return
	}

	HandleGetBook(w, r)
}

func HandleDeleteBook(w http.ResponseWriter, r *http.Request) {
	bookid := strings.Replace(r.URL.Path, "/book/", "", 1)
	err := bookStore.DelBook(bookid)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		error, _ := json.Marshal(fmt.Sprintf("%v", err))

		w.Write(error)

		return
	}

	HandleGetBooks(w, r)
}

// BOOK

type Book struct {
	Id     string `json:"id"`
	Author string `json:"author"`
	Name   string `json:"name"`
}

type BookStore struct {
	books []Book
}

var bookStore = BookStore{
	books: make([]Book, 0),
}

func (s BookStore) GetBooks() []Book {
	return s.books
}

func (s BookStore) FindBookById(id string) *Book {
	for _, book := range s.books {
		if book.Id == id {
			return &book
		}
	}

	return nil
}

func (s *BookStore) AddBook(book Book) error {
	bk := s.FindBookById(book.Id)
	if bk != nil {
		return errors.New(fmt.Sprintf("Book with id %s already exists", book.Id))

	}
	s.books = append(s.books, book)

	return nil
}

func (s *BookStore) SetBook(book Book) error {
	for i, bk := range s.books {
		if bk.Id == book.Id {

			s.books[i] = book

			return nil
		}
	}

	return errors.New(fmt.Sprintf("There is no book with id %s", book.Id))
}

func (s *BookStore) DelBook(id string) error {
	for i, bk := range s.books {
		if bk.Id == id {

			s.books = append(s.books[:i], s.books[i+1:]...)

			return nil
		}
	}

	return errors.New(fmt.Sprintf("There is no book with id %s", id))
}
