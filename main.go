package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	nats "github.com/nats-io/go-nats"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"

	"github.com/gorilla/mux"
)

var (
	natsURL  = "demo.nats.io"           // Can be superseded by env NATSURL
	natsPort = ":4222"                  // Can be superseded by env NATSPORT
	natsPost = "zjnO12CgNkHD0IsuGd89zA" // POST new post channel. Can be superseded by env NATSPOST
	natsGet  = "OWM7pKQNbXd7l75l21kOzA" // GET posts channel. Can be superseded by env NATSGET
	port     = ":8080"
	nc       *nats.Conn
	wg       = sync.WaitGroup{}
)

// Message is the representation of a post
type blog struct {
	ID      string
	Title   string
	Content string
	Date    string
}

// superTrace keeps jaeger ctx
type superTrace struct {
	TraceID map[string]string
}

type cookieStatus struct {
	Cookie string
}

func initJaeger(service string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}
	tracer, closer, err := cfg.New(service, config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}

// serveBlogs displays all the blogs registered in the database
func serveBlogs(w http.ResponseWriter, r *http.Request) {
	pages := getPages()
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Println(err)
	}
	err = t.Execute(w, pages)
	if err != nil {
		log.Println(err)
	}
}

// getPages makes a request through NATS to get all records from the database
func getPages() []blog {

	tracer, closer := initJaeger("getPages")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	span := tracer.StartSpan("GetRaleurFront")

	carrierVar := make(map[string]string)
	span.Tracer().Inject(span.Context(), opentracing.TextMap, opentracing.TextMapCarrier(carrierVar))

	trace := superTrace{}
	trace.TraceID = carrierVar

	traceByte, err := json.Marshal(trace)

	pages := []blog{}
	if os.Getenv("NATSGET") != "" {
		natsGet = os.Getenv("NATSGET")
	}
	msg, err := nc.Request(natsGet, traceByte, time.Second*3)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("get records on %s\n", natsGet)
	}
	json.Unmarshal(msg.Data, &pages)

	span.Finish()

	return pages
}

// newBlog displays a blank form to create a new blog
func newBlog(w http.ResponseWriter, r *http.Request) {
	thisPage := blog{}
	t, err := template.ParseFiles("templates/new.html")
	if err != nil {
		log.Println(err)
	}
	err = t.Execute(w, thisPage)
	if err != nil {
		log.Println(err)
	}
}

// newBlogStatus retrieves the return code from the cookie
// and informs about the status of the blog registration
func newBlogStatus(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/status.html")
	if err != nil {
		log.Println(err)
	}
	err = t.Execute(w, getStatus(r))
	if err != nil {
		log.Println(err)
	}
}

// getStatus retrieves and returns the return code contained in the cookie
func getStatus(r *http.Request) cookieStatus {
	cookieValue, err := r.Cookie("message")
	if err != nil {
		log.Println(err)
	}
	textDecode, err := url.QueryUnescape(cookieValue.Value)
	if err != nil {
		log.Println(err)
	}
	return cookieStatus{Cookie: textDecode}
}

// postBlog sends data from a new blog to the backend
func postBlog(w http.ResponseWriter, r *http.Request) {
	var (
		msg         = blog{}
		cookieValue string
	)
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

	// The status of sending the blog to the backend is kept in a cookie
	if os.Getenv("NATSPOST") != "" {
		natsPost = os.Getenv("NATSPOST")
	}
	err = nc.Publish(natsPost, post)
	if err != nil {
		log.Println(err)
		cookieValue = "Quelque chose de terrible s'est produit..."
	} else {
		log.Println("new POST published on NATS...")
		textEncode := &url.URL{Path: "Le coup de gueule a été envoyé !..."}
		cookieValue = textEncode.String()
	}
	cookie := http.Cookie{Name: "message", Value: cookieValue, Expires: time.Now().Add(3 * time.Second), HttpOnly: true}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/api/status", 301)
}

func main() {

	if os.Getenv("NATSURL") != "" {
		natsURL = os.Getenv("NATSURL")
	}
	if os.Getenv("NATSPORT") != "" {
		natsPort = os.Getenv("NATSPORT")
	}

	// TODO: this should be a flag
	if os.Getenv("HS-MICRO-FRONT") != "" {
		port = os.Getenv("HS-MICRO-FRONT")
	}

	go func() {
		var err error
		nc, err = nats.Connect("nats://" + natsURL + natsPort)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("connected to nats://%s%s\n", natsURL, natsPort)
		}
		defer nc.Close()
		wg.Add(1)
		wg.Wait()
	}()

	rtr := mux.NewRouter()
	rtr.HandleFunc("/", serveBlogs).Methods("GET")
	rtr.HandleFunc("/new", newBlog).Methods("GET")
	rtr.HandleFunc("/api/status", newBlogStatus).Methods("GET")
	rtr.HandleFunc("/api/post", postBlog).Methods("POST")
	rtr.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", rtr)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}

}
