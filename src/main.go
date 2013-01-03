// Package main implements a (prototype-quality) review wizard.
//
// The code is organized into three packages:
//
//    entity
//      ^   ^
//      |    \
//      |    persist
//      |     ^
//      |    /
//      main
//
// The entity package contains domain syntax: Reviews, (Question) Profiles, Responses, etc.
// The persist package implements the *Repo interfaces defined in the entity model.
// The main package defines the web UI and calls the entity and persist code.
//
// For more information on this architecture, see
//
//     http://manuel.kiessling.net/2012/09/28/applying-the-clean-architecture-to-go-applications/
//
package main

import (
	"entity"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"html"
	"html/template"
	"log"
	"net/http"
	"persist"
	"runtime/debug"
	"time"

// revel, pat, gorilla
// gorp, json, xml
)

type App struct {
	*mux.Router
	entity.ProfileRepo
	entity.ReviewRepo
}

func recoverHTTP(w http.ResponseWriter, r *http.Request) {
	if rec := recover(); rec != nil {
		switch err := rec.(type) {
		case error:
			log.Printf("error: %v, req: %v", err, r)
			debug.PrintStack()
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			log.Printf("unknown error: %v, req: %v", err, r)
			debug.PrintStack()
			http.Error(w, "unknown error", http.StatusInternalServerError)
		}
	}
}

func checkHTTP(err error) {
	if err != nil {
		panic(err)
	}
}

func HandleReviewSetGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleReviewSetGet()\n")
	// ...
	// list links to all (current?) reviews?
	reviews, err := self.GetAllReviews()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "review_set", reviews)
}

// BUG(mistone): CSRF!
func HandleReviewSetPost(self *App, w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/reviews/" {
		http.Error(w,
			fmt.Sprintf("Bad Path %q %q\n",
				html.EscapeString(r.Method),
				html.EscapeString(r.URL.Path)),
			http.StatusBadRequest)
	} else {
		// parse body
		reviewName := r.FormValue("review")

		reviewVer, err := entity.NewVersionFromString(reviewName)
		checkHTTP(err)

		log.Printf("HandleReviewSetPost(): reviewVer: %v\n", reviewVer)

		// extract profile
		profileName := r.FormValue("profile")

		profileVer, err := entity.NewVersionFromString(profileName)
		checkHTTP(err)

		profile, err := self.GetProfileById(*profileVer)
		checkHTTP(err)

		log.Printf("HandleReviewSetPost(): profile: %v\n", profile)

		// make a new Review
		review := &entity.Review{
			Version:   *reviewVer,
			Profile:   profile,
			Responses: make([]*entity.Response, len(profile.Questions)),
		}

		// make appropriate Responses based on the Questions
		//   contained in the indicated Profile
		for idx, q := range profile.Questions {
			review.Responses[idx] = &entity.Response{
				Question: q,
				Answer:   nil,
			}
		}
		log.Printf("HandleReviewSetPost(): created review: %v\n", review)

		// persist the new review
		err = self.ReviewRepo.AddReview(review)
		checkHTTP(err)

		url, err := self.Router.Get("review").URL("review_name", reviewVer.String())
		checkHTTP(err)

		log.Printf("HandleReviewSetPost(): redirecting to: %v\n", url)
		http.Redirect(w, r, url.String(), http.StatusSeeOther)
	}
}

func HandleReviewGet(self *App, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reviewName := vars["review_name"]
	log.Printf("HandleReviewGet(): reviewName: %v\n", reviewName)

	reviewVer, err := entity.NewVersionFromString(reviewName)
	checkHTTP(err)
	log.Printf("HandleReviewGet(): reviewVer: %v\n", reviewVer)

	review, err := self.ReviewRepo.GetReviewById(*reviewVer)
	checkHTTP(err)
	log.Printf("HandleReviewGet(): review: %v\n", review)

	renderTemplate(w, "review", review)
}

func HandleReviewPost(self *App, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reviewName := vars["review_name"]
	log.Printf("HandleReviewPost(): reviewName: %v\n", reviewName)

	reviewVer, err := entity.NewVersionFromString(reviewName)
	checkHTTP(err)
	log.Printf("HandleReviewPost(): reviewVer: %v\n", reviewVer)

	review, err := self.ReviewRepo.GetReviewById(*reviewVer)
	checkHTTP(err)
	log.Printf("HandleReviewPost(): review: %v\n", review)

	questionName := r.FormValue("question_name")
	log.Printf("HandleReviewPost(): questionName: %v\n", questionName)

	questionVer, err := entity.NewVersionFromString(questionName)
	checkHTTP(err)
	log.Printf("HandleReviewPost(): questionVer: %v\n", questionVer)

	datum := r.FormValue(questionVer.String())
	log.Printf("HandleReviewPost(): datum: %v\n", datum)

	answer := &entity.Answer{
		Author:       "", // BUG(mistone): need to set author!
		CreationTime: time.Now(),
		Datum:        datum,
	}
	review, err = review.SetResponseAnswer(*questionVer, answer)
	checkHTTP(err)

	err = self.AddReview(review)
	checkHTTP(err)

	log.Printf("HandleReviewPost(): done\n")
	url, err := self.Router.Get("review").URL("review_name", reviewVer.String())
	checkHTTP(err)
	url.Fragment = "response-" + questionVer.String()

	log.Printf("HandleReviewPost(): redirecting to: %v\n", url)
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

func HandleProfileSetGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleProfileSetGet()\n")
	// ...
	// list links to all (current?) profiles?
	profiles, err := self.GetAllProfiles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "profile_set", profiles)
}

func HandleProfileSetPost(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleProfileSetPost()\n")
	http.Error(w, "Not implemented.", http.StatusNotImplemented)
}

func HandleProfileGet(self *App, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	profileName := vars["profile_name"]
	log.Printf("HandleProfileGet(): profileName: %v\n", profileName)

	profileVer, err := entity.NewVersionFromString(profileName)
	checkHTTP(err)
	log.Printf("HandleProfileGet(): profileVer: %v\n", profileVer)

	profile, err := self.ProfileRepo.GetProfileById(*profileVer)
	checkHTTP(err)
	log.Printf("HandleProfileGet(): profile: %v\n", profile)

	renderTemplate(w, "profile", profile)
}

func HandleProfilePost(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleProfilePost()\n")
	http.Error(w, "Not implemented.", http.StatusNotImplemented)
}

func HandleRootGet(self *App, w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "root", nil)
}

var templates = template.Must(template.ParseFiles(
	"src/root.html",
	"src/review.html",
	"src/review_set.html",
	"src/profile.html",
	"src/profile_set.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var prefill = flag.Bool("prefill", false, "prefill data.json")

func doServe() {
	persist := persist.NewPersistJSON()

	r := mux.NewRouter()

	app := &App{
		ProfileRepo: entity.ProfileRepo(persist),
		ReviewRepo:  entity.ReviewRepo(persist),
		Router:      r,
	}

	wrap := func(fn func(*App, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return (func(w http.ResponseWriter, r *http.Request) {
			defer recoverHTTP(w, r)
			fn(app, w, r)
		})
	}

	fmt.Printf("Hi.\n")

	s := r.PathPrefix("/reviews").Subrouter()
	s.HandleFunc("/", wrap(HandleReviewSetGet)).Methods("GET").Name("review_set")
	s.HandleFunc("/", wrap(HandleReviewSetPost)).Methods("POST")
	s.HandleFunc("/{review_name}", wrap(HandleReviewGet)).Methods("GET").Name("review")
	s.HandleFunc("/{review_name}", wrap(HandleReviewPost)).Methods("POST")

	s = r.PathPrefix("/profiles").Subrouter()
	s.HandleFunc("/", wrap(HandleProfileSetGet)).Methods("GET").Name("profile_set")
	s.HandleFunc("/", wrap(HandleProfileSetPost)).Methods("POST")
	s.HandleFunc("/{profile_name}", wrap(HandleProfileGet)).Methods("GET").Name("profile")
	s.HandleFunc("/{profile_name}", wrap(HandleProfilePost)).Methods("POST")

	r.HandleFunc("/", wrap(HandleRootGet)).Methods("GET").Name("root")

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe("127.0.0.1:3001", nil))
}

func doPrefill() {
	pMem := &persist.PersistMem{}
	pJS := persist.NewPersistJSON()

	profiles, err := pMem.GetAllProfiles()
	if err != nil {
		panic(err)
	}
	for _, prof := range profiles {
		err = pJS.AddProfile(prof)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	flag.Parse()

	if *prefill {
		doPrefill()
	} else {
		doServe()
	}
}
