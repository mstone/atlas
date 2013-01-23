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
	"bytes"
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
	"sort"
	"strings"
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

// BUG(mistone): vProfileList and vReviewList sorting should use version sorts, not string sorts
type vProfile struct {
	*entity.Profile
	Selected bool
}
type vProfileList []vProfile

func (s vProfileList) Len() int { return len(s) }
func (s vProfileList) Less(i, j int) bool {
	return s[i].Version.String() < s[j].Version.String()
}
func (s vProfileList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type vReviewList []*entity.Review

func (s vReviewList) Len() int { return len(s) }
func (s vReviewList) Less(i, j int) bool {
	return s[i].Version.String() < s[j].Version.String()
}
func (s vReviewList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type vReviewSet struct {
	Profiles vProfileList
	Reviews  vReviewList
}

func HandleReviewSetGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleReviewSetGet()\n")

	profiles, err := self.GetAllProfiles()
	checkHTTP(err)

	// ...
	// list links to all (current?) reviews?
	reviews, err := self.GetAllReviews()
	checkHTTP(err)

	profilesList := make(vProfileList, len(profiles))
	for idx, prof := range profiles {
		isPace := prof.Version.String() == "pace-1.0.0"
		profilesList[idx] = vProfile{
			Profile:  prof,
			Selected: isPace,
		}
	}
	sort.Sort(profilesList)

	reviewsList := vReviewList(reviews)
	sort.Sort(reviewsList)

	view := &vReviewSet{
		Profiles: profilesList,
		Reviews:  reviewsList,
	}

	renderTemplate(w, "review_set", view)
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
			Responses: make(map[entity.Version]*entity.Response, len(profile.Questions)),
		}

		// make appropriate Responses based on the Questions
		//   contained in the indicated Profile
		for idx, q := range profile.Questions {
			review.Responses[idx] = &entity.Response{
				Question: q,
				Answer:   entity.NewAnswer(),
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

type vResponse struct {
	*entity.Response
	AnswerInput template.HTML
	DatumList   []string
}

type vResponseList []*vResponse

func (s vResponseList) Len() int { return len(s) }
func (s vResponseList) Less(i, j int) bool {
	if s[i].SortKey == s[j].SortKey {
		return s[i].Question.Version.Name < s[j].Question.Version.Name
	}
	return s[i].SortKey < s[j].SortKey
}
func (s vResponseList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type vResponseGroup struct {
	GroupKey     string
	ResponseList vResponseList
}

type vResponseGroupList []*vResponseGroup

func (s vResponseGroupList) Len() int           { return len(s) }
func (s vResponseGroupList) Less(i, j int) bool { return s[i].GroupKey < s[j].GroupKey }
func (s vResponseGroupList) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type vReview struct {
	ReviewName     string
	ProfileName    string
	ResponseNames  []string
	ResponseGroups vResponseGroupList
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

	// produce sorted groups of sorted responses
	responseGroupsMap := make(map[string]vResponseList)
	for _, resp := range review.Responses {
		log.Printf("HandleReviewGet(): considering resp %v", resp)

		vresp := &vResponse{
			Response: resp,
		}

		var templateName string
		switch resp.Question.AnswerType {
		default:
			panic(fmt.Sprintf("Unknown answer type for question: %s", resp.Question.Version))
		case 0:
			templateName = "textarea.html"
		case 1:
			templateName = "multiselect.html"
			vresp.DatumList = strings.Split(resp.Answer.Datum, " ")
		}

		var buf bytes.Buffer
		err = templates.ExecuteTemplate(&buf, templateName, vresp)
		checkHTTP(err)

		vresp.AnswerInput = template.HTML(buf.String())

		responseGroupsMap[resp.Question.GroupKey] = append(responseGroupsMap[vresp.Question.GroupKey], vresp)
	}
	log.Printf("HandleReviewGet(): produced responseGroupsMap %v", responseGroupsMap)
	responseGroupsList := make(vResponseGroupList, len(responseGroupsMap))
	counter := 0
	for groupKey, respList := range responseGroupsMap {
		sort.Sort(respList)
		responseGroup := &vResponseGroup{
			GroupKey:     groupKey,
			ResponseList: respList,
		}
		responseGroupsList[counter] = responseGroup
		counter++
	}
	log.Printf("HandleReviewGet(): produced responseGroupsList %v", responseGroupsList)
	log.Printf("HandleReviewGet(): sorting responseGroupsList", responseGroupsList)
	sort.Sort(responseGroupsList)
	log.Printf("HandleReviewGet(): got final responseGroupsList", responseGroupsList)

	responseNames := getProfileQuestionNames(review.Profile)

	view := vReview{
		ReviewName:     review.Version.String(),
		ProfileName:    review.Profile.Version.String(),
		ResponseNames:  responseNames,
		ResponseGroups: responseGroupsList,
	}
	log.Printf("HandleReviewGet(): view: %v\n", view)

	// render view
	renderTemplate(w, "review", view)
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
	checkHTTP(err)
	renderTemplate(w, "profile_set", profiles)
}

func HandleProfileSetPost(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleProfileSetPost()\n")

	// extract profile
	profileName := r.FormValue("profile")
	profileVer, err := entity.NewVersionFromString(profileName)
	checkHTTP(err)

	// check for old profile
	oldProfile, err := self.GetProfileById(*profileVer)
	if oldProfile != nil {
		http.Error(w, "Profile already exists.", http.StatusConflict)
	}

	// make a new Profile
	newProfile := &entity.Profile{
		Version: *profileVer,
	}

	// persist the new review
	err = self.ProfileRepo.AddProfile(newProfile)
	checkHTTP(err)

	url, err := self.Router.Get("profile").URL("profile_name", profileVer.String())
	checkHTTP(err)

	log.Printf("HandleProfileSetPost(): redirecting to: %v\n", url)
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

type vQuestionList []*entity.Question

func (s vQuestionList) Len() int           { return len(s) }
func (s vQuestionList) Less(i, j int) bool { return s[i].SortKey < s[j].SortKey }
func (s vQuestionList) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type vQuestionGroup struct {
	GroupKey  string
	Questions vQuestionList
}

type vQuestionGroupList []*vQuestionGroup

func (s vQuestionGroupList) Len() int           { return len(s) }
func (s vQuestionGroupList) Less(i, j int) bool { return s[i].GroupKey < s[j].GroupKey }
func (s vQuestionGroupList) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type vProfile2 struct {
	ProfileName    string
	QuestionNames  []string
	QuestionGroups vQuestionGroupList
}

func getProfileQuestionNames(profile *entity.Profile) []string {
	questionNames := make([]string, len(profile.Questions))
	idx := 0
	for k := range profile.Questions {
		questionNames[idx] = k.String()
		idx++
	}
	sort.Strings(questionNames)
	return questionNames
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

	// produce sorted groups of sorted questions
	questionGroupMap := make(map[string]vQuestionList)
	for _, quest := range profile.Questions {
		log.Printf("HandleProfileGet(): considering question %v", quest)
		questionGroupMap[quest.GroupKey] = append(questionGroupMap[quest.GroupKey], quest)
	}
	log.Printf("HandleProfileGet(): produced questionGroupMap %v", questionGroupMap)
	questionGroupList := make(vQuestionGroupList, len(questionGroupMap))
	counter := 0
	for groupKey, questList := range questionGroupMap {
		sort.Sort(questList)
		questionGroup := &vQuestionGroup{
			GroupKey:  groupKey,
			Questions: questList,
		}
		questionGroupList[counter] = questionGroup
		counter++
	}
	log.Printf("HandleProfileGet(): produced questionGroupList %v", questionGroupList)
	log.Printf("HandleProfileGet(): sorting questionGroupList", questionGroupList)
	sort.Sort(questionGroupList)
	log.Printf("HandleProfileGet(): got final questionGroupList", questionGroupList)

	questionNames := getProfileQuestionNames(profile)

	view := vProfile2{
		ProfileName:    profile.Version.String(),
		QuestionNames:  questionNames,
		QuestionGroups: questionGroupList,
	}
	log.Printf("HandleProfileGet(): view: %v\n", view)

	renderTemplate(w, "profile", view)
}

func HandleProfilePost(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleProfilePost()\n")
	http.Error(w, "Not implemented.", http.StatusNotImplemented)
}

type vQuestion struct {
	QuestionName string
}

type vQuestionSet struct {
	QuestionNames []string
}

func getAllQuestions(self *App) ([]*entity.Question, error) {
	profiles, err := self.GetAllProfiles()
	checkHTTP(err)

	questionMap := map[string]*entity.Question{}
	for _, prof := range profiles {
		for k, v := range prof.Questions {
			// BUG(mistone): check for duplicates?
			questionMap[k.String()] = v
		}
	}

	questions := make([]*entity.Question, len(questionMap))
	idx := 0
	for _, v := range questionMap {
		questions[idx] = v
		idx++
	}
	return questions, nil
}

func getAllQuestionNames(self *App) ([]string, error) {
	questions, err := getAllQuestions(self)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(questions))
	for k, v := range questions {
		names[k] = v.Version.String()
	}

	return names, nil
}

func HandleQuestionSetGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleQuestionSetGet()\n")

	names, err := getAllQuestionNames(self)
	checkHTTP(err)

	view := vQuestionSet{
		QuestionNames: names,
	}
	log.Printf("HandleQuestionSetGet(): view: %v", view)
	sort.Strings(view.QuestionNames)

	renderTemplate(w, "question_set", view)
}

func HandleQuestionSetPost(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleQuestionSetPost()\n")
	http.Error(w, "Not implemented.", http.StatusNotImplemented)
}

func HandleQuestionGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleQuestionGet()\n")

	vars := mux.Vars(r)
	questionName := vars["question_name"]
	log.Printf("HandleQuestionGet(): questionName: %v\n", questionName)

	questionVer, err := entity.NewVersionFromString(questionName)
	checkHTTP(err)
	log.Printf("HandleQuestionGet(): questionVer: %v\n", questionVer)

	view := &vQuestion{
		QuestionName: questionVer.String(),
	}
	renderTemplate(w, "question", view)
}

func HandleQuestionPost(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleQuestionPost()\n")
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
	"src/profile_set.html",
	"src/question.html",
	"src/question_set.html",
	"src/textarea.html",
	"src/multiselect.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

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

	s = r.PathPrefix("/questions").Subrouter()
	s.HandleFunc("/", wrap(HandleQuestionSetGet)).Methods("GET").Name("question_set")
	s.HandleFunc("/", wrap(HandleQuestionSetPost)).Methods("POST")
	s.HandleFunc("/{question_name}", wrap(HandleQuestionGet)).Methods("GET").Name("question")
	s.HandleFunc("/{question_name}", wrap(HandleQuestionPost)).Methods("POST")

	r.HandleFunc("/", wrap(HandleRootGet)).Methods("GET").Name("root")

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe("127.0.0.1:3001", nil))
}

func main() {
	flag.Parse()

	doServe()
}
