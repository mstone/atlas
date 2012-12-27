package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"time"
	//"regexp"
	"html/template"

//"go/build"
// html "github.com/moovweb/gokogiri/html"
// _ "github.com/nsf/gocode"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

type Profile struct {
	Id        int
	Version   Version
	Questions []*Question
}

type Question struct {
	Id          int
	Version     Version
	Label       []byte
	Text        []byte
	Help        []byte
	AnswerType  int
	DisplayHint []byte
}

type QuestionDep struct {
	Id   int
	From *Question
	To   []*Question
}

type Answer struct {
	Id           int
	Question     *Question
	Datum        []byte
	CreationTime time.Time
	Author       []byte
	// DescribedTime
}

type QuestionAnswer struct {
	Question
	Answer
}

type AnswerProfile struct {
	Id        int
	Version   Version
	Questions []*QuestionAnswer
}

type Review struct {
	Id            int
	AnswerProfile *AnswerProfile
}

var idChan chan int
var rootProfile Profile
var rootReviews []Review

func handleReviewSetPost(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/reviews/" {
		http.Error(w,
			fmt.Sprintf("Bad Path %q %q\n",
					html.EscapeString(r.Method),
					html.EscapeString(r.URL.Path)),
			http.StatusBadRequest)
	} else {
		fmt.Fprintf(w, "Hello, %q\n", html.EscapeString(r.URL.Path))
		id := <-idChan
		fmt.Fprintf(w, "Next ID: %d\n", id)
	}
}

func handleReviewSetGet(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "reviews", nil)
}

func handleReviewSet(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		handleReviewSetPost(w, r)
	case "GET":
		handleReviewSetGet(w, r)
	default:
		http.Error(w,
			fmt.Sprintf("Unknown Method %q %q\n",
					html.EscapeString(r.Method),
					html.EscapeString(r.URL.Path)),
			http.StatusMethodNotAllowed)
	}
}

var templates = template.Must(template.ParseFiles("src/reviews.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// Create a supply of ids.
	idChan = make(chan int)
	go func() {
		for i := 0; ; i++ {
			idChan <- i
		}
	}()
	fmt.Printf("Hi.\n")
	http.HandleFunc("/reviews/", handleReviewSet)
	log.Fatal(http.ListenAndServe("127.0.0.1:3001", nil))
}
