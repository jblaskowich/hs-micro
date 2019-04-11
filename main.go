package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	nats "github.com/nats-io/go-nats"

	"github.com/gorilla/mux"
)

var (
	cookieValue string
)

type Message struct {
	Title   string
	Content string
}

type Page struct {
	Cookie string
}

func servePages(w http.ResponseWriter, r *http.Request) {
	thisPage := Message{}
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, thisPage)
}

func displayStatus(w http.ResponseWriter, r *http.Request) {
	thisPage := Page{}
	cookieVal, _ := r.Cookie("message")
	thisPage.Cookie = cookieVal.Value
	t, _ := template.ParseFiles("templates/status.html")
	t.Execute(w, thisPage)
}

func postMessage(w http.ResponseWriter, r *http.Request) {
	msg := Message{}
	err := r.ParseForm()
	msg.Title = r.FormValue("title")
	msg.Content = r.FormValue("content")
	post, err := json.Marshal(msg)
	if err != nil {
		log.Println(err.Error())
	}

	natsURL := os.Getenv("NATSURL")
	if natsURL == "" {
		natsURL = "demo.nats.io"
	}
	natsPort := os.Getenv("NATSPORT")
	if natsPort == "" {
		natsPort = ":4222"
	}
	natsChan := os.Getenv("NATSCHAN")
	if natsChan == "" {
		natsChan = "zjnO12CgNkHD0IsuGd89zA"
	}

	nc, err := nats.Connect("nats://" + natsURL + natsPort)
	if err != nil {
		log.Println(err.Error())
	}
	err = nc.Publish(natsChan, post)
	if err != nil {
		log.Println(err.Error())
		cookieValue = "Something wrong happends..."
	} else {
		cookieValue = "The post was sent!..."
	}
	cookie := http.Cookie{Name: "message", Value: cookieValue, Expires: time.Now().Add(3 * time.Second), HttpOnly: true}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/api/status", 301)
}

func viewMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello viewMessage")
}

func main() {
	port := os.Getenv("HS-MICRO-FRONT")
	if port == "" {
		port = ":8080"
	}
	rtr := mux.NewRouter()
	rtr.HandleFunc("/", servePages).Methods("GET")
	rtr.HandleFunc("/api/post", postMessage).Methods("POST")
	rtr.HandleFunc("/api/status", displayStatus).Methods("GET")
	rtr.HandleFunc("/api/view", viewMessage).Methods("GET")
	http.Handle("/", rtr)
	http.ListenAndServe(port, nil)
}
