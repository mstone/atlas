package main

import (
	"entity"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"persist"

//"go/build"
// html "github.com/moovweb/gokogiri/html"
// _ "github.com/nsf/gocode"
// revel, pat, gorilla
// gorp, json, xml
)

type App struct {
	idChan chan int
	entity.ProfileRepo
	entity.QuestionRepo
	entity.ReviewRepo
}

type ReviewSetApp struct {
	*App
}

type RootApp struct {
	*App
}

type ProfileSetApp struct {
	*App
}

func (self *ReviewSetApp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer recoverHTTP(w, r)
	self.HandleReviewSet(w, r)
}

func (self *ProfileSetApp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer recoverHTTP(w, r)
	self.HandleProfileSet(w, r)
}

func (self *RootApp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer recoverHTTP(w, r)
	self.HandleRoot(w, r)
}

func recoverHTTP(w http.ResponseWriter, r *http.Request) {
	if rec := recover(); rec != nil {
		switch err := rec.(type) {
		case error:
			log.Printf("error: %v, req: %v", err, r)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			log.Printf("unknown error: %v, req: %v", err, r)
			http.Error(w, "unknown error", http.StatusInternalServerError)
		}
	}
}

func checkHTTP(err error) {
	if err != nil {
		panic(err)
	}
}

func (self *App) HandleReviewSetPost(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/reviews/" {
		http.Error(w,
			fmt.Sprintf("Bad Path %q %q\n",
				html.EscapeString(r.Method),
				html.EscapeString(r.URL.Path)),
			http.StatusBadRequest)
	} else {
		// ...
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

		// get a new id
		// make a new Review
		review := &entity.Review{
			Version: *reviewVer,
			Responses: make([]*entity.Response, len(profile.Questions)),
		}

		// make appropriate QuestionAnswers based on the Questions
		//   contained in the indicated Profile
		for idx, q := range profile.Questions {
			review.Responses[idx] = &entity.Response{
				Question: q,
				Answer: nil,
			}
		}
		log.Printf("HandleReviewSetPost(): created review: %v\n", review)
		// persist the new review
		err = self.ReviewRepo.AddReview(review)
		checkHTTP(err)
	}
}

func (self *App) HandleReviewSetGet(w http.ResponseWriter, r *http.Request) {
	// ...
	// list links to all (current?) reviews?
	reviews, err := self.GetAllReviews()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "reviews", reviews)
}

func (self *App) HandleReviewSet(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		self.HandleReviewSetPost(w, r)
	case "GET":
		self.HandleReviewSetGet(w, r)
	default:
		http.Error(w,
			fmt.Sprintf("Unknown Method %q %q\n",
				html.EscapeString(r.Method),
				html.EscapeString(r.URL.Path)),
			http.StatusMethodNotAllowed)
	}
}

func (self *App) HandleProfileSetPost(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/profiles/" {
		http.Error(w,
			fmt.Sprintf("Bad Path %q %q\n",
				html.EscapeString(r.Method),
				html.EscapeString(r.URL.Path)),
			http.StatusBadRequest)
	} else {
		// ...
		// parse body
		// extract && look up Profile
		// get a new id
		// make a new Review
		// make a new AnswerProfile
		// make appropriate QuestionAnswers based on the Questions
		//   contained in the indicated Profile
		// add the new Review to rootReviews
		// save
		fmt.Fprintf(w, "Hello, %q\n", html.EscapeString(r.URL.Path))
		id := <-self.idChan
		fmt.Fprintf(w, "Next ID: %d\n", id)
	}
}

func (self *App) HandleProfileSetGet(w http.ResponseWriter, r *http.Request) {
	// ...
	// list links to all (current?) profiles?
	profiles, err := self.GetAllProfiles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "profiles", profiles)
}

func (self *App) HandleProfileSet(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		self.HandleProfileSetPost(w, r)
	case "GET":
		self.HandleProfileSetGet(w, r)
	default:
		http.Error(w,
			fmt.Sprintf("Unknown Method %q %q\n",
				html.EscapeString(r.Method),
				html.EscapeString(r.URL.Path)),
			http.StatusMethodNotAllowed)
	}
}

func (self *App) HandleRootGet(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "root", nil)
}

func (self *App) HandleRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		self.HandleRootGet(w, r)
	default:
		http.Error(w,
			fmt.Sprintf("Unknown Method %q %q\n",
				html.EscapeString(r.Method),
				html.EscapeString(r.URL.Path)),
			http.StatusMethodNotAllowed)
	}

}

var templates = template.Must(template.ParseFiles(
	"src/reviews.html",
	"src/root.html",
	"src/profiles.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// Create a supply of ids.
	idChan := make(chan int)
	go func() {
		for i := 0; ; i++ {
			idChan <- i
		}
	}()

	persist := new(persist.PersistMem)

	app := &App{
		idChan:       idChan,
		ProfileRepo:  entity.ProfileRepo(persist),
		QuestionRepo: entity.QuestionRepo(persist),
		ReviewRepo:   entity.ReviewRepo(persist),
	}

	fmt.Printf("Hi.\n")
	http.Handle("/reviews/",
		&ReviewSetApp{app})
	http.Handle("/profiles/",
		&ProfileSetApp{app})
	http.Handle("/",
		&RootApp{app})
	log.Fatal(http.ListenAndServe("127.0.0.1:3001", nil))
}
