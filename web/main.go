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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/russross/blackfriday"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
)

const MAX_CHART_SIZE = 1000000

type App struct {
	entity.QuestionRepo
	entity.ProfileRepo
	entity.ReviewRepo
	StaticPath string
	StaticRoot string
	FormsRoot  string
	ChartsRoot string
	HtmlPath   string
	ChartsPath string
	HttpAddr   string
	//router     *mux.Router
	templates *template.Template
	//wrap       func(fn func(*App, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request)
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

func (self *App) GetReviewUrl(version *entity.Version) (url.URL, error) {
	url := url.URL{
		Path: path.Clean(path.Join(self.FormsRoot, "reviews", version.String())),
	}
	return url, nil
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
		reviewUrl, err := self.GetReviewUrl(&review.Version)
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
		vRoot:    newVRoot(self, "review_set", "", "", ""),
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

	url, err := self.GetReviewUrl(reviewVer)
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

func (self *App) ParseUrlReviewVersion(r *http.Request) (*entity.Version, error) {
	reviewName := r.URL.Path[len(path.Clean(path.Join(self.FormsRoot, "reviews"))+"/"):]
	log.Printf("ParseUrlReviewVersion(): reviewName: %v\n", reviewName)

	reviewVer, err := entity.NewVersionFromString(reviewName)
	checkHTTP(err)
	log.Printf("ParseUrlReviewVersion(): reviewVer: %v\n", reviewVer)

	return reviewVer, err
}

func HandleReviewGet(self *App, w http.ResponseWriter, r *http.Request) {
	reviewVer, err := self.ParseUrlReviewVersion(r)
	checkHTTP(err)
	log.Printf("HandleReviewGet(): reviewVer: %v\n", reviewVer)

	review, err := self.ReviewRepo.GetReviewById(*reviewVer)
	checkHTTP(err)
	log.Printf("HandleReviewGet(): review: %v\n", review)

	// produce sorted groups of sorted responses
	responseGroupsMap := make(map[string]vResponseList)
	for _, resp := range review.Responses {
		log.Printf("HandleReviewGet(): considering resp %v", resp)

		questionUrl, err := self.GetQuestionUrl(&resp.Question.Version)
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
		vRoot:          newVRoot(self, "review", "", "", ""),
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
	reviewVer, err := self.ParseUrlReviewVersion(r)
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

	url, err := self.GetReviewUrl(reviewVer)
	checkHTTP(err)
	url.Fragment = "response-" + questionVer.String()

	log.Printf("HandleReviewPost(): redirecting to: %v\n", url)
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

type vProfile3 struct {
	*entity.Profile
	Url url.URL
}

type vProfileSet struct {
	*vRoot
	Profiles []*vProfile3
}

func HandleProfileSetGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleProfileSetGet()\n")
	// ...
	// list links to all (current?) profiles?
	profiles, err := self.GetAllProfiles()
	checkHTTP(err)

	vProfiles := make([]*vProfile3, len(profiles))
	for idx, profile := range profiles {
		url, err := self.GetProfileUrl(&profile.Version)
		checkHTTP(err)
		vProfiles[idx] = &vProfile3{
			Profile: profile,
			Url:     url,
		}
	}

	view := &vProfileSet{
		vRoot:    newVRoot(self, "profile_set", "", "", ""),
		Profiles: vProfiles,
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

	url, err := self.GetProfileUrl(profileVer)
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

func (self *App) GetQuestionUrl(version *entity.Version) (url.URL, error) {
	url := url.URL{
		Path: path.Clean(path.Join(self.FormsRoot, "questions", version.String())),
	}
	return url, nil
}

func HandleProfileGet(self *App, w http.ResponseWriter, r *http.Request) {
	profileVer, err := self.ParseUrlProfileVersion(r)
	checkHTTP(err)
	log.Printf("HandleProfileGet(): profileVer: %v\n", profileVer)

	profile, err := self.ProfileRepo.GetProfileById(*profileVer)
	checkHTTP(err)
	log.Printf("HandleProfileGet(): profile: %v\n", profile)

	// produce sorted groups of sorted questions
	questionGroupMap := make(map[string]vQuestionList)
	for _, quest := range profile.Questions {
		log.Printf("HandleProfileGet(): considering question %v", quest)

		questionUrl, err := self.GetQuestionUrl(&quest.Version)
		checkHTTP(err)
		log.Printf("HandleProfileGet(): questionUrl: %v\n", questionUrl)

		vQuest := vQuestion{
			vRoot:        newVRoot(self, "question", "", "", ""),
			Url:          questionUrl,
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
		vRoot:          newVRoot(self, "profile", "", "", ""),
		ProfileName:    profile.Version.String(),
		QuestionNames:  questionNames,
		QuestionGroups: questionGroupList,
	}
	log.Printf("HandleProfileGet(): view: %v\n", view)

	self.renderTemplate(w, "profile", view)
}

func (self *App) GetProfileUrl(version *entity.Version) (url.URL, error) {
	url := url.URL{
		Path: path.Clean(path.Join(self.FormsRoot, "profiles", version.String())),
	}
	return url, nil
}

func (self *App) ParseUrlProfileVersion(r *http.Request) (*entity.Version, error) {
	profileName := r.URL.Path[len(path.Clean(path.Join(self.FormsRoot, "profiles"))+"/"):]
	log.Printf("ParseUrlProfileVersion(): profileName: %v\n", profileName)

	profileVer, err := entity.NewVersionFromString(profileName)
	checkHTTP(err)
	log.Printf("ParseUrlProfileVersion(): profileVer: %v\n", profileVer)

	return profileVer, err
}

func HandleProfilePost(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleProfilePost()\n")

	profileVer, err := self.ParseUrlProfileVersion(r)
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

	url, err := self.GetProfileUrl(profileVer)
	checkHTTP(err)
	url.Fragment = "question-" + questionVer.String()

	log.Printf("HandleProfilePost(): redirecting to: %v\n", url)
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

type vQuestion struct {
	*vRoot
	QuestionName string
	Url          url.URL
	*entity.Question
}

type vQuestionSet struct {
	*vRoot
	Questions vQuestion2List
}

type vQuestion2 struct {
	Name string
	Url  url.URL
}

type vQuestion2List []*vQuestion2

func (s vQuestion2List) Len() int { return len(s) }
func (s vQuestion2List) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}
func (s vQuestion2List) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func getAllQuestionLinks(self *App) ([]*vQuestion2, error) {
	questions, err := self.GetAllQuestions()
	if err != nil {
		return nil, err
	}

	view := make([]*vQuestion2, len(questions))
	for idx, question := range questions {
		url, err := self.GetQuestionUrl(&question.Version)
		checkHTTP(err)
		view[idx] = &vQuestion2{
			Name: question.Version.String(),
			Url:  url,
		}
	}

	return view, nil
}

func HandleQuestionSetGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleQuestionSetGet()\n")

	links, err := getAllQuestionLinks(self)
	checkHTTP(err)

	view := vQuestionSet{
		vRoot:     newVRoot(self, "question_set", "", "", ""),
		Questions: links,
	}
	log.Printf("HandleQuestionSetGet(): view: %v", view)
	sort.Sort(view.Questions)

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

	url, err := self.GetQuestionUrl(questionVer)
	checkHTTP(err)

	log.Printf("HandleQuestionSetPost(): redirecting to: %v\n", url)
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

func (self *App) ParseUrlQuestionVersion(r *http.Request) (*entity.Version, error) {
	questionName := r.URL.Path[len(path.Clean(path.Join(self.FormsRoot, "questions"))+"/"):]
	log.Printf("ParseUrlQuestionVersion(): questionName: %v\n", questionName)

	questionVer, err := entity.NewVersionFromString(questionName)
	checkHTTP(err)
	log.Printf("ParseUrlQuestionVersion(): questionVer: %v\n", questionVer)

	return questionVer, err
}

func HandleQuestionGet(self *App, w http.ResponseWriter, r *http.Request) {
	questionVer, err := self.ParseUrlQuestionVersion(r)
	checkHTTP(err)
	log.Printf("HandleQuestionGet(): questionVer: %v\n", questionVer)

	question, err := self.GetQuestionById(*questionVer)
	checkHTTP(err)
	log.Printf("HandleQuestionGet(): question: %v\n", question)

	questionUrl, err := self.GetQuestionUrl(questionVer)
	checkHTTP(err)
	log.Printf("HandleQuestionGet(): questionUrl: %v\n", questionUrl)

	view := &vQuestion{
		vRoot:        newVRoot(self, "question", "", "", ""),
		Url:          questionUrl,
		QuestionName: questionVer.String(),
		Question:     question,
	}
	self.renderTemplate(w, "question", view)
}

func HandleQuestionPost(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleQuestionPost()\n")

	questionVer, err := self.ParseUrlQuestionVersion(r)
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

	url, err := self.GetQuestionUrl(questionVer)
	checkHTTP(err)

	log.Printf("HandleQuestionPost(): redirecting to: %v\n", url)
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

type vChart struct {
	*vRoot
	FullPath string
	Url      string
	Html     template.HTML
}

func (self *App) ReadChart(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size := fi.Size()
	if size > MAX_CHART_SIZE {
		return nil, errors.New(fmt.Sprintf("Chart too big: %s", path))
	}

	buf := make([]byte, size)

	n, err := f.Read(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[0:n]

	return buf, nil
}

func HandleChartGet(self *App, w http.ResponseWriter, r *http.Request) {
	chartUrl := path.Clean(r.URL.Path)
	log.Printf("HandleChartGet(): chartUrl: %v\n", chartUrl)

	fullPath := path.Join(self.ChartsPath, chartUrl)

	if chartUrl == "/site.json" {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleSiteJsonGet(self, w, r)
		}
		return
	}

	buf := make([]byte, 1000000)

	// anyway, assuming it's a chart, find the index.txt

	fi, err := os.Stat(fullPath)
	checkHTTP(err)

	if !fi.IsDir() {
		fp3 := fullPath
		// BUG(mistone): don't set Content-Type blindly; also need to check Accept header
		// BUG(mistone): do we really want to sniff mime-types here?
		http.ServeFile(w, r, fp3)
		return
	} else {
		fp1 := path.Join(fullPath, "index.txt")
		f, err := os.Open(fp1)
		if err != nil {
			if os.IsNotExist(err) {
				fp2 := path.Join(fullPath, "index.text")
				f2, err := os.Open(fp2)
				checkHTTP(err)
				defer f2.Close()

				n, err := f2.Read(buf)
				checkHTTP(err)
				buf = buf[0:n]
			} else {
				panic(err)
			}
		} else {
			defer f.Close()
			n, err := f.Read(buf)
			checkHTTP(err)
			buf = buf[0:n]
		}

		// attempt to parse header lines
		title := ""
		authors := ""
		date := ""
		sbuf := string(buf)
		lines := strings.Split(sbuf, "\n")
		log.Printf("HandleChartGet: found %d lines", len(lines))
		if len(lines) > 3 {
			title = strings.TrimLeft(lines[0], "% ")
			authors = strings.TrimLeft(lines[1], "% ")
			date = strings.TrimLeft(lines[2], "% ")
			rest := strings.SplitAfterN(sbuf, "\n", 4)
			buf = []byte(rest[3])
		}
		log.Printf("HandleChartGet: found title: %s", title)
		log.Printf("HandleChartGet: found authors: %s", authors)
		log.Printf("HandleChartGet: found date: %s", date)

		htmlFlags := 0
		//htmlFlags |= blackfriday.HTML_USE_XHTML
		htmlFlags |= blackfriday.HTML_TOC

		renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

		extFlags := 0
		extFlags |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
		extFlags |= blackfriday.EXTENSION_TABLES
		extFlags |= blackfriday.EXTENSION_FENCED_CODE
		extFlags |= blackfriday.EXTENSION_AUTOLINK
		extFlags |= blackfriday.EXTENSION_STRIKETHROUGH
		extFlags |= blackfriday.EXTENSION_SPACE_HEADERS

		html := blackfriday.Markdown(buf, renderer, extFlags)

		view := &vChart{
			vRoot: newVRoot(self, "chart", title, authors, date),
			//Url:          chartUrl.String(),
			FullPath: fullPath,
			Url:      chartUrl,
			Html:     template.HTML(html),
		}

		self.renderTemplate(w, "chart", view)
	}
}

type vRoot struct {
	PageName       string
	Title          string
	Authors        string
	Date           string
	StaticUrl      string
	FormsRoot      string
	ChartsRoot     string
	QuestionSetUrl string
	ProfileSetUrl  string
	ReviewSetUrl   string
}

func newVRoot(self *App, pageName string, title string, authors string, date string) *vRoot {
	questionSetUrl := url.URL{Path: path.Clean(path.Join(self.FormsRoot, "questions"))}
	profileSetUrl := url.URL{Path: path.Clean(path.Join(self.FormsRoot, "profiles"))}
	reviewSetUrl := url.URL{Path: path.Clean(path.Join(self.FormsRoot, "reviews"))}

	return &vRoot{
		PageName:       pageName,
		Title:          title,
		Authors:        authors,
		Date:           date,
		StaticUrl:      self.StaticRoot,
		FormsRoot:      self.FormsRoot,
		ChartsRoot:     path.Clean(self.ChartsRoot + "/"),
		QuestionSetUrl: questionSetUrl.String(),
		ProfileSetUrl:  profileSetUrl.String(),
		ReviewSetUrl:   reviewSetUrl.String(),
	}
}

func HandleRootGet(self *App, w http.ResponseWriter, r *http.Request) {
	view := newVRoot(self, "root", "", "", "")
	self.renderTemplate(w, "root", view)
}

func HandleSiteJsonGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleSiteJsonGet(): start")

	view := map[string]string{}
	filepath.Walk(self.ChartsPath, func(name string, fi os.FileInfo, err error) error {
		log.Printf("HandleSiteJsonGet(): visiting path %s", name)
		if err != nil {
			return err
		}

		dir := filepath.Dir(name)
		base := filepath.Base(name)
		pfx := path.Clean(self.ChartsPath)

		var sfx string
		if len(dir) > len(pfx) {
			sfx = dir[len(pfx)+1:] + "/"
		} else {
			sfx = ""
		}
		key := sfx

		log.Printf("HandleSiteJsonGet(): pfx %s", pfx)
		log.Printf("HandleSiteJsonGet(): dir %s", dir)
		log.Printf("HandleSiteJsonGet(): base %s", base)
		log.Printf("HandleSiteJsonGet(): sfx %s", sfx)

		switch base {
		default:
			return nil
		case "index.text":
			fallthrough
		case "index.txt":
			text, err := self.ReadChart(name)
			if err != nil {
				log.Printf("HandleSiteJsonGet(): warning %s", err)
			} else {
				view[key] = string(text)
			}
		}
		return nil
	})

	writer := bufio.NewWriter(w)
	defer writer.Flush()

	encoder := json.NewEncoder(writer)
	err := encoder.Encode(&view)
	checkHTTP(err)

	//log.Printf("SiteJsonGet(): encoded view: %v", view)
}

func (self *App) renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := self.templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (self *App) HandleStatic(w http.ResponseWriter, r *http.Request) {
	// BUG(mistone): directory traversal?
	up := path.Clean(r.URL.Path)
	sp := path.Clean("/"+self.StaticRoot) + "/"
	fp := up[len(sp):]
	log.Printf("HandleStatic: file path: %v", fp)
	http.ServeFile(w, r, path.Join(self.StaticPath, fp))
}

func (self *App) HandleForms(w http.ResponseWriter, r *http.Request) {
	up := path.Clean(r.URL.Path)
	sp := path.Clean("/" + self.FormsRoot)
	fp := up[len(sp):]

	log.Printf("HandleForms: file path: %v", fp)

	if fp == "" {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleRootGet(self, w, r)
		}
		return
	}

	if fp == "/reviews" {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleReviewSetGet(self, w, r)
		case "POST":
			HandleReviewSetPost(self, w, r)
		}
		return
	}

	if fp == "/questions" {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleQuestionSetGet(self, w, r)
		case "POST":
			HandleQuestionSetPost(self, w, r)
		}
		return
	}

	if fp == "/profiles" {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleProfileSetGet(self, w, r)
		case "POST":
			HandleProfileSetPost(self, w, r)
		}
		return
	}

	if strings.HasPrefix(fp, "/reviews/") {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleReviewGet(self, w, r)
		case "POST":
			HandleReviewPost(self, w, r)
		}
		return
	}

	if strings.HasPrefix(fp, "/questions/") {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleQuestionGet(self, w, r)
		case "POST":
			HandleQuestionPost(self, w, r)
		}
		return
	}

	if strings.HasPrefix(fp, "/profiles/") {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleProfileGet(self, w, r)
		case "POST":
			HandleProfilePost(self, w, r)
		}
		return
	}

	panic("unknown form resource")
}

func (self *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer recoverHTTP(w, r)
	log.Printf("HandleRootApp: path: %v", r.URL.Path)
	isStatic := strings.HasPrefix(r.URL.Path, path.Clean("/"+self.StaticRoot))
	if isStatic {
		log.Printf("HandleRootApp: dispatching to static...")
		self.HandleStatic(w, r)
	} else {
		isForm := strings.HasPrefix(r.URL.Path, path.Clean("/"+self.FormsRoot))
		if isForm {
			log.Printf("HandleRootApp: dispatching to forms...")
			self.HandleForms(w, r)
			return
		} else {
			isChart := strings.HasPrefix(r.URL.Path, path.Clean("/"+self.ChartsRoot))
			if isChart {
				log.Printf("HandleRootApp: dispatching to charts...")
				HandleChartGet(self, w, r)
			} else {
				panic(fmt.Sprintf("Can't route path: %v", r.URL.Path))
			}
		}
	}
}

// Serve initializes some variables on self and then delegates to net/http to
// to receive incoming HTTP requests. Requests are handled by self.ServeHTTP()
func (self *App) Serve() {
	self.templates = template.Must(
		template.ParseGlob(
			path.Join(self.HtmlPath, "*.html")))

	self.StaticRoot = path.Clean("/" + self.StaticRoot)
	self.FormsRoot = path.Clean("/" + self.FormsRoot)

	fmt.Printf("App: %v\n", self)

	http.Handle("/", self)
	log.Fatal(http.ListenAndServe(self.HttpAddr, nil))
}
