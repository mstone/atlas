// Package web implements the review wizard's controllers
// and views.
//
// Presently, there are controllers for these resources:
//
//     Root
//       QuestionSet
//         Question
//       ProfileSet
//         Profile
//       ReviewSet
//         Review
//
// Controllers are App struct methods named Handle*{Get|Post}.
//
// Controllers largely work by
//
//   1. looking up an entity to be modified or represented in the requested
//      response
//
//   2. constructing a private "view struct" with the data to be displayed
//
//   3. handing the view struct and the http.ResponseWriter to an appropriate
//      template for rendering.
//
// All *Set resources load and save their contained entities through
// corresponding *Repo interfaces on the App struct.
package web

import (
	"akamai/atlas/forms/entity"
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"path"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
)

type App struct {
	entity.QuestionRepo
	entity.ProfileRepo
	entity.ReviewRepo
	StaticUrl  string
	AppRoot    string
	HtmlPath   string
	StaticPath string
	HttpAddr   string
	router     *mux.Router
	templates  *template.Template
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

type vReview2 struct {
	Url string
	*entity.Review
}

type vReviewList []vReview2

func (s vReviewList) Len() int { return len(s) }
func (s vReviewList) Less(i, j int) bool {
	return s[i].Version.String() < s[j].Version.String()
}
func (s vReviewList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type vReviewSet struct {
	*vRoot
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

	reviewsList := make(vReviewList, len(reviews))
	for idx, review := range reviews {
		reviewUrl, err := self.router.Get("review").URL("review_name", review.Version.String())
		checkHTTP(err)

		reviewsList[idx] = vReview2{
			Url:    reviewUrl.String(),
			Review: review,
		}
	}
	sort.Sort(reviewsList)

	view := &vReviewSet{
		Profiles: profilesList,
		Reviews:  reviewsList,
		vRoot:    newVRoot(self, "review_set"),
	}

	self.renderTemplate(w, "review_set", view)
}

// BUG(mistone): CSRF!
func HandleReviewSetPost(self *App, w http.ResponseWriter, r *http.Request) {
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

	url, err := self.router.Get("review").URL("review_name", reviewVer.String())
	checkHTTP(err)

	log.Printf("HandleReviewSetPost(): redirecting to: %v\n", url)
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

type vResponse struct {
	*entity.Response
	AnswerInput template.HTML
	DatumList   []string
	QuestionUrl string
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
	*vRoot
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

		questionUrl, err := self.router.Get("question").URL("question_name", resp.Question.Version.String())
		checkHTTP(err)
		vresp := &vResponse{
			Response:    resp,
			QuestionUrl: questionUrl.String(),
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
		err = self.templates.ExecuteTemplate(&buf, templateName, vresp)
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
		vRoot:          newVRoot(self, "review"),
		ReviewName:     review.Version.String(),
		ProfileName:    review.Profile.Version.String(),
		ResponseNames:  responseNames,
		ResponseGroups: responseGroupsList,
	}
	log.Printf("HandleReviewGet(): view: %v\n", view)

	// render view
	self.renderTemplate(w, "review", view)
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
	url, err := self.router.Get("review").URL("review_name", reviewVer.String())
	checkHTTP(err)
	url.Fragment = "response-" + questionVer.String()

	log.Printf("HandleReviewPost(): redirecting to: %v\n", url)
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

type vProfileSet struct {
	*vRoot
	Profiles []*entity.Profile
}

func HandleProfileSetGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleProfileSetGet()\n")
	// ...
	// list links to all (current?) profiles?
	profiles, err := self.GetAllProfiles()
	checkHTTP(err)

	view := &vProfileSet{
		vRoot:    newVRoot(self, "profile_set"),
		Profiles: profiles,
	}
	self.renderTemplate(w, "profile_set", view)
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

	url, err := self.router.Get("profile").URL("profile_name", profileVer.String())
	checkHTTP(err)

	log.Printf("HandleProfileSetPost(): redirecting to: %v\n", url)
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

type vQuestionList []vQuestion

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
	*vRoot
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

		questionUrl, err := self.router.Get("question").URL("question_name", quest.Version.String())
		checkHTTP(err)
		log.Printf("HandleProfileGet(): questionUrl: %v\n", questionUrl)

		vQuest := vQuestion{
			vRoot:        newVRoot(self, "question"),
			Url:          questionUrl.String(),
			QuestionName: quest.Version.String(),
			Question:     quest,
		}

		questionGroupMap[quest.GroupKey] = append(questionGroupMap[quest.GroupKey], vQuest)
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
		vRoot:          newVRoot(self, "profile"),
		ProfileName:    profile.Version.String(),
		QuestionNames:  questionNames,
		QuestionGroups: questionGroupList,
	}
	log.Printf("HandleProfileGet(): view: %v\n", view)

	self.renderTemplate(w, "profile", view)
}

func HandleProfilePost(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleProfilePost()\n")

	vars := mux.Vars(r)
	profileName := vars["profile_name"]
	log.Printf("HandleProfilePost(): profileName: %v\n", profileName)

	profileVer, err := entity.NewVersionFromString(profileName)
	checkHTTP(err)
	log.Printf("HandleProfilePost(): profileVer: %v\n", profileVer)

	profile, err := self.ProfileRepo.GetProfileById(*profileVer)
	checkHTTP(err)
	log.Printf("HandleProfilePost(): profile: %v\n", profile)

	questionName := r.FormValue("question_name")
	log.Printf("HandleProfilePost(): questionName: %v\n", questionName)

	questionVer, err := entity.NewVersionFromString(questionName)
	checkHTTP(err)
	log.Printf("HandleProfilePost(): questionVer: %v\n", questionVer)

	question := &entity.Question{
		Version: *questionVer,
	}
	oldQuestion, err := self.GetQuestionById(*questionVer)
	if oldQuestion != nil {
		question = oldQuestion
	} else {
		log.Printf("HandleProfilePost(): generating fresh question: %v\n", question)
		err = self.QuestionRepo.AddQuestion(question)
		checkHTTP(err)
	}

	profile.Questions[*questionVer] = question
	err = self.AddProfile(profile)
	checkHTTP(err)

	url, err := self.router.Get("profile").URL("profile_name", profileVer.String())
	checkHTTP(err)
	url.Fragment = "question-" + questionVer.String()

	log.Printf("HandleProfilePost(): redirecting to: %v\n", url)
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

type vQuestion struct {
	*vRoot
	QuestionName string
	Url          string
	*entity.Question
}

type vQuestionSet struct {
	*vRoot
	QuestionNames []string
}

func getAllQuestionNames(self *App) ([]string, error) {
	questions, err := self.GetAllQuestions()
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
		vRoot:         newVRoot(self, "question_set"),
		QuestionNames: names,
	}
	log.Printf("HandleQuestionSetGet(): view: %v", view)
	sort.Strings(view.QuestionNames)

	self.renderTemplate(w, "question_set", view)
}

func HandleQuestionSetPost(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleQuestionSetPost()\n")

	questionName := r.FormValue("question_name")
	log.Printf("HandleQuestionSetPost(): questionName: %v\n", questionName)

	questionVer, err := entity.NewVersionFromString(questionName)
	checkHTTP(err)
	log.Printf("HandleQuestionSetPost(): questionVer: %v\n", questionVer)

	oldQuestion, err := self.GetQuestionById(*questionVer)
	if oldQuestion != nil {
		http.Error(w, "Question already exists.", http.StatusConflict)
	}

	question := &entity.Question{
		Version: *questionVer,
	}
	log.Printf("HandleQuestionSetPost(): question: %v\n", question)

	err = self.QuestionRepo.AddQuestion(question)
	checkHTTP(err)

	url, err := self.router.Get("question").URL("question_name", questionVer.String())
	checkHTTP(err)

	log.Printf("HandleQuestionSetPost(): redirecting to: %v\n", url)
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

func HandleQuestionGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleQuestionGet()\n")

	vars := mux.Vars(r)
	questionName := vars["question_name"]
	log.Printf("HandleQuestionGet(): questionName: %v\n", questionName)

	questionVer, err := entity.NewVersionFromString(questionName)
	checkHTTP(err)
	log.Printf("HandleQuestionGet(): questionVer: %v\n", questionVer)

	question, err := self.GetQuestionById(*questionVer)
	checkHTTP(err)
	log.Printf("HandleQuestionGet(): question: %v\n", question)

	questionUrl, err := self.router.Get("question").URL("question_name", questionVer.String())
	checkHTTP(err)
	log.Printf("HandleQuestionGet(): questionUrl: %v\n", questionUrl)

	view := &vQuestion{
		vRoot:        newVRoot(self, "question"),
		Url:          questionUrl.String(),
		QuestionName: questionVer.String(),
		Question:     question,
	}
	self.renderTemplate(w, "question", view)
}

func HandleQuestionPost(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleQuestionPost()\n")

	vars := mux.Vars(r)
	questionName := vars["question_name"]
	log.Printf("HandleQuestionPost(): questionName: %v\n", questionName)

	questionVer, err := entity.NewVersionFromString(questionName)
	checkHTTP(err)
	log.Printf("HandleQuestionPost(): questionVer: %v\n", questionVer)

	question, err := self.QuestionRepo.GetQuestionById(*questionVer)
	checkHTTP(err)
	log.Printf("HandleQuestionPost(): question: %v\n", question)

	question.Text = r.FormValue("question_text")
	question.Help = r.FormValue("question_help")
	question.DisplayHint = r.FormValue("question_displayhint")
	question.GroupKey = r.FormValue("question_groupkey")
	sortkey, err := strconv.Atoi(r.FormValue("question_sortkey"))
	checkHTTP(err)
	question.SortKey = sortkey

	err = self.QuestionRepo.AddQuestion(question)
	checkHTTP(err)

	url, err := self.router.Get("question").URL("question_name", questionVer.String())
	checkHTTP(err)

	log.Printf("HandleQuestionPost(): redirecting to: %v\n", url)
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

type vRoot struct {
	PageName       string
	StaticUrl      string
	AppRoot        string
	QuestionSetUrl string
	ProfileSetUrl  string
	ReviewSetUrl   string
}

func newVRoot(self *App, pageName string) *vRoot {
	questionSetUrl, _ := self.router.Get("question_set").URL()
	profileSetUrl, _ := self.router.Get("profile_set").URL()
	reviewSetUrl, _ := self.router.Get("review_set").URL()

	return &vRoot{
		PageName:       pageName,
		StaticUrl:      self.StaticUrl,
		AppRoot:        self.AppRoot,
		QuestionSetUrl: questionSetUrl.String(),
		ProfileSetUrl:  profileSetUrl.String(),
		ReviewSetUrl:   reviewSetUrl.String(),
	}
}

func HandleRootGet(self *App, w http.ResponseWriter, r *http.Request) {
	view := newVRoot(self, "root")
	self.renderTemplate(w, "root", view)
}

func (self *App) renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := self.templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Serve loads HTML templates from self.HtmlPath, connects wrapped controllers
// to a gorilla/mux request router, and then uses net/http to receive and to act
// on incoming HTTP requests.
func (self *App) Serve() {
	rr := mux.NewRouter()
	var r *mux.Router
	var staticUrl string
	if self.AppRoot != "" {
		r = rr.PathPrefix("/" + self.AppRoot).Subrouter()
		staticUrl = "/" + self.AppRoot + "/static/"
	} else {
		r = rr
		staticUrl = "/static/"
	}
	fmt.Printf("Using appRoot: /%s\n", self.AppRoot)
	fmt.Printf("Using staticUrl: %s\n", staticUrl)
	self.router = r

	self.templates = template.Must(
		template.ParseGlob(
			path.Join(self.HtmlPath, "*.html")))

	self.StaticUrl = path.Clean(staticUrl)

	wrap := func(fn func(*App, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return (func(w http.ResponseWriter, r *http.Request) {
			defer recoverHTTP(w, r)
			fn(self, w, r)
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

	http.Handle(staticUrl, http.StripPrefix(staticUrl, http.FileServer(http.Dir(self.StaticPath))))
	http.Handle("/", rr)
	log.Fatal(http.ListenAndServe(self.HttpAddr, nil))
}
