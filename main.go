package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	nats "github.com/nats-io/go-nats"

	"github.com/gorilla/mux"
)

var (
	cookieValue string
	natsURL     = "demo.nats.io"           // Can be superseded by env NATSURL
	natsPort    = ":4222"                  // Can be superseded by env NATSPORT
	natsPost    = "zjnO12CgNkHD0IsuGd89zA" // Can be superseded by env NATSCPOST
	natsGet     = "OWM7pKQNbXd7l75l21kOzA" // Can be superseded by env NATSGET
)

// Message is the representation of a post
type Message struct {
	ID      string
	Title   string
	Content string
	Date    string
}

// State is used to capture the status of the message sent to nats
type State struct {
	Cookie string
}

func servePages(w http.ResponseWriter, r *http.Request) {
	Page := []Message{}

	if os.Getenv("NATSURL") != "" {
		natsURL = os.Getenv("NATSURL")
	}
	if os.Getenv("NATSPORT") != "" {
		natsPort = os.Getenv("NATSPORT")
	}
	if os.Getenv("NATSGET") != "" {
		natsGet = os.Getenv("NATSGET")
	}

	nc, err := nats.Connect("nats://" + natsURL + natsPort)
	if err != nil {
		log.Println(err.Error())
	}
	defer nc.Close()

	// This request will generate an inbox for the backend to reply
	msg, err := nc.Request(natsGet, nil, time.Second*60)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(msg.Data, &Page)

	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, Page)
}

func newMessage(w http.ResponseWriter, r *http.Request) {
	thisPage := Message{}
	t, _ := template.ParseFiles("templates/new.html")
	t.Execute(w, thisPage)
}

func displayStatus(w http.ResponseWriter, r *http.Request) {
	thisPage := State{}
	cookieVal, _ := r.Cookie("message")
	textDecode, _ := url.QueryUnescape(cookieVal.Value)
	thisPage.Cookie = textDecode
	t, _ := template.ParseFiles("templates/status.html")
	t.Execute(w, thisPage)
}

func postMessage(w http.ResponseWriter, r *http.Request) {
	msg := Message{}
	err := r.ParseForm()
	if err != nil {
		log.Println(err.Error())
	}
	msg.Title = r.FormValue("title")
	msg.Content = r.FormValue("content")
	post, err := json.Marshal(msg)
	if err != nil {
		log.Println(err.Error())
	}

	if os.Getenv("NATSURL") != "" {
		natsURL = os.Getenv("NATSURL")
	}
	if os.Getenv("NATSPORT") != "" {
		natsPort = os.Getenv("NATSPORT")
	}
	if os.Getenv("NATSPOST") != "" {
		natsPost = os.Getenv("NATSPOST")
	}

	nc, err := nats.Connect("nats://" + natsURL + natsPort)
	if err != nil {
		log.Println(err.Error())
	}
	err = nc.Publish(natsPost, post)
	if err != nil {
		log.Println(err.Error())
		cookieValue = "Quelque chose de terrible s'est produit..."
	} else {
		textEncode := &url.URL{Path: "Le coup de gueule a été envoyé !..."}
		cookieValue = textEncode.String()
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
	rtr.HandleFunc("/new", newMessage).Methods("GET")
	rtr.HandleFunc("/api/status", displayStatus).Methods("GET")
	rtr.HandleFunc("/api/post", postMessage).Methods("POST")
	rtr.HandleFunc("/api/view", viewMessage).Methods("GET")
	http.Handle("/", rtr)
	http.ListenAndServe(port, nil)
}
