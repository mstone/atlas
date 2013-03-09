// Package web implements the record wizard's controllers
// and views.
//
// Presently, there are controllers for these resources:
//
//     Root
//       QuestionSet
//         Question
//       FormSet
//         Form
//       RecordSet
//         Record
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
	"akamai/atlas/forms/chart"
	"akamai/atlas/forms/entity"
	"akamai/atlas/forms/linker"
	"bufio"
	"bytes"
	"encoding/json"
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
	entity.FormRepo
	entity.RecordRepo
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

// BUG(mistone): vFormList and vRecordList sorting should use version sorts, not string sorts
type vForm struct {
	*entity.Form
	Selected bool
}
type vFormList []vForm

func (s vFormList) Len() int { return len(s) }
func (s vFormList) Less(i, j int) bool {
	return s[i].Version.String() < s[j].Version.String()
}
func (s vFormList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type vRecord2 struct {
	Url string
	*entity.Record
}

type vRecordList []vRecord2

func (s vRecordList) Len() int { return len(s) }
func (s vRecordList) Less(i, j int) bool {
	return s[i].Version.String() < s[j].Version.String()
}
func (s vRecordList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type vRecordSet struct {
	*vRoot
	Forms   vFormList
	Records vRecordList
}

func (self *App) GetRecordUrl(version *entity.Version) (url.URL, error) {
	url := url.URL{
		Path: path.Clean(path.Join(self.FormsRoot, "records", version.String())),
	}
	return url, nil
}

func HandleRecordSetGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleRecordSetGet()\n")

	forms, err := self.GetAllForms()
	checkHTTP(err)

	// ...
	// list links to all (current?) records?
	records, err := self.GetAllRecords()
	checkHTTP(err)

	formsList := make(vFormList, len(forms))
	for idx, prof := range forms {
		isPace := prof.Version.String() == "pace-1.0.0"
		formsList[idx] = vForm{
			Form:     prof,
			Selected: isPace,
		}
	}
	sort.Sort(formsList)

	recordsList := make(vRecordList, len(records))
	for idx, record := range records {
		recordUrl, err := self.GetRecordUrl(&record.Version)
		checkHTTP(err)

		recordsList[idx] = vRecord2{
			Url:    recordUrl.String(),
			Record: record,
		}
	}
	sort.Sort(recordsList)

	view := &vRecordSet{
		Forms:   formsList,
		Records: recordsList,
		vRoot:   newVRoot(self, "record_set", "", "", ""),
	}

	self.renderTemplate(w, "record_set", view)
}

// BUG(mistone): CSRF!
func HandleRecordSetPost(self *App, w http.ResponseWriter, r *http.Request) {
	// parse body
	recordName := r.FormValue("record")

	recordVer, err := entity.NewVersionFromString(recordName)
	checkHTTP(err)

	log.Printf("HandleRecordSetPost(): recordVer: %v\n", recordVer)

	// extract form
	formName := r.FormValue("form")

	formVer, err := entity.NewVersionFromString(formName)
	checkHTTP(err)

	form, err := self.GetFormById(*formVer)
	checkHTTP(err)

	log.Printf("HandleRecordSetPost(): form: %v\n", form)

	// make a new Record
	record := &entity.Record{
		Version:   *recordVer,
		Form:      form,
		Responses: make(map[entity.Version]*entity.Response, len(form.Questions)),
	}

	// make appropriate Responses based on the Questions
	//   contained in the indicated Form
	for idx, q := range form.Questions {
		record.Responses[idx] = &entity.Response{
			Question: q,
			Answer:   entity.NewAnswer(),
		}
	}
	log.Printf("HandleRecordSetPost(): created record: %v\n", record)

	// persist the new record
	err = self.RecordRepo.AddRecord(record)
	checkHTTP(err)

	url, err := self.GetRecordUrl(recordVer)
	checkHTTP(err)

	log.Printf("HandleRecordSetPost(): redirecting to: %v\n", url)
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

type vRecord struct {
	*vRoot
	RecordName     string
	FormName       string
	ResponseNames  []string
	ResponseGroups vResponseGroupList
}

func (self *App) ParseUrlRecordVersion(r *http.Request) (*entity.Version, error) {
	recordName := r.URL.Path[len(path.Clean(path.Join(self.FormsRoot, "records"))+"/"):]
	log.Printf("ParseUrlRecordVersion(): recordName: %v\n", recordName)

	recordVer, err := entity.NewVersionFromString(recordName)
	checkHTTP(err)
	log.Printf("ParseUrlRecordVersion(): recordVer: %v\n", recordVer)

	return recordVer, err
}

func HandleRecordGet(self *App, w http.ResponseWriter, r *http.Request) {
	recordVer, err := self.ParseUrlRecordVersion(r)
	checkHTTP(err)
	log.Printf("HandleRecordGet(): recordVer: %v\n", recordVer)

	record, err := self.RecordRepo.GetRecordById(*recordVer)
	checkHTTP(err)
	log.Printf("HandleRecordGet(): record: %v\n", record)

	// produce sorted groups of sorted responses
	responseGroupsMap := make(map[string]vResponseList)
	for _, resp := range record.Responses {
		log.Printf("HandleRecordGet(): considering resp %v", resp)

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
	log.Printf("HandleRecordGet(): produced responseGroupsMap %v", responseGroupsMap)
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
	log.Printf("HandleRecordGet(): produced responseGroupsList %v", responseGroupsList)
	log.Printf("HandleRecordGet(): sorting responseGroupsList", responseGroupsList)
	sort.Sort(responseGroupsList)
	log.Printf("HandleRecordGet(): got final responseGroupsList", responseGroupsList)

	responseNames := getFormQuestionNames(record.Form)

	view := vRecord{
		vRoot:          newVRoot(self, "record", "", "", ""),
		RecordName:     record.Version.String(),
		FormName:       record.Form.Version.String(),
		ResponseNames:  responseNames,
		ResponseGroups: responseGroupsList,
	}
	log.Printf("HandleRecordGet(): view: %v\n", view)

	// render view
	self.renderTemplate(w, "record", view)
}

func HandleRecordPost(self *App, w http.ResponseWriter, r *http.Request) {
	recordVer, err := self.ParseUrlRecordVersion(r)
	checkHTTP(err)
	log.Printf("HandleRecordPost(): recordVer: %v\n", recordVer)

	record, err := self.RecordRepo.GetRecordById(*recordVer)
	checkHTTP(err)
	log.Printf("HandleRecordPost(): record: %v\n", record)

	questionName := r.FormValue("question_name")
	log.Printf("HandleRecordPost(): questionName: %v\n", questionName)

	questionVer, err := entity.NewVersionFromString(questionName)
	checkHTTP(err)
	log.Printf("HandleRecordPost(): questionVer: %v\n", questionVer)

	datum := r.FormValue(questionVer.String())
	log.Printf("HandleRecordPost(): datum: %v\n", datum)

	answer := &entity.Answer{
		Author:       "", // BUG(mistone): need to set author!
		CreationTime: time.Now(),
		Datum:        datum,
	}
	record, err = record.SetResponseAnswer(*questionVer, answer)
	checkHTTP(err)

	err = self.AddRecord(record)
	checkHTTP(err)

	log.Printf("HandleRecordPost(): done\n")

	url, err := self.GetRecordUrl(recordVer)
	checkHTTP(err)
	url.Fragment = "response-" + questionVer.String()

	log.Printf("HandleRecordPost(): redirecting to: %v\n", url)
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

type vForm3 struct {
	*entity.Form
	Url url.URL
}

type vFormSet struct {
	*vRoot
	Forms []*vForm3
}

func HandleFormSetGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleFormSetGet()\n")
	// ...
	// list links to all (current?) forms?
	forms, err := self.GetAllForms()
	checkHTTP(err)

	vForms := make([]*vForm3, len(forms))
	for idx, form := range forms {
		url, err := self.GetFormUrl(&form.Version)
		checkHTTP(err)
		vForms[idx] = &vForm3{
			Form: form,
			Url:  url,
		}
	}

	view := &vFormSet{
		vRoot: newVRoot(self, "form_set", "", "", ""),
		Forms: vForms,
	}
	self.renderTemplate(w, "form_set", view)
}

func HandleFormSetPost(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleFormSetPost()\n")

	// extract form
	formName := r.FormValue("form")
	formVer, err := entity.NewVersionFromString(formName)
	checkHTTP(err)

	// check for old form
	oldForm, err := self.GetFormById(*formVer)
	if oldForm != nil {
		http.Error(w, "Form already exists.", http.StatusConflict)
	}

	// make a new Form
	newForm := &entity.Form{
		Version: *formVer,
	}

	// persist the new record
	err = self.FormRepo.AddForm(newForm)
	checkHTTP(err)

	url, err := self.GetFormUrl(formVer)
	checkHTTP(err)

	log.Printf("HandleFormSetPost(): redirecting to: %v\n", url)
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

type vForm2 struct {
	*vRoot
	FormName       string
	QuestionNames  []string
	QuestionGroups vQuestionGroupList
}

func getFormQuestionNames(form *entity.Form) []string {
	questionNames := make([]string, len(form.Questions))
	idx := 0
	for k := range form.Questions {
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

func HandleFormGet(self *App, w http.ResponseWriter, r *http.Request) {
	formVer, err := self.ParseUrlFormVersion(r)
	checkHTTP(err)
	log.Printf("HandleFormGet(): formVer: %v\n", formVer)

	form, err := self.FormRepo.GetFormById(*formVer)
	checkHTTP(err)
	log.Printf("HandleFormGet(): form: %v\n", form)

	// produce sorted groups of sorted questions
	questionGroupMap := make(map[string]vQuestionList)
	for _, quest := range form.Questions {
		log.Printf("HandleFormGet(): considering question %v", quest)

		questionUrl, err := self.GetQuestionUrl(&quest.Version)
		checkHTTP(err)
		log.Printf("HandleFormGet(): questionUrl: %v\n", questionUrl)

		vQuest := vQuestion{
			vRoot:        newVRoot(self, "question", "", "", ""),
			Url:          questionUrl,
			QuestionName: quest.Version.String(),
			Question:     quest,
		}

		questionGroupMap[quest.GroupKey] = append(questionGroupMap[quest.GroupKey], vQuest)
	}
	log.Printf("HandleFormGet(): produced questionGroupMap %v", questionGroupMap)
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
	log.Printf("HandleFormGet(): produced questionGroupList %v", questionGroupList)
	log.Printf("HandleFormGet(): sorting questionGroupList", questionGroupList)
	sort.Sort(questionGroupList)
	log.Printf("HandleFormGet(): got final questionGroupList", questionGroupList)

	questionNames := getFormQuestionNames(form)

	view := vForm2{
		vRoot:          newVRoot(self, "form", "", "", ""),
		FormName:       form.Version.String(),
		QuestionNames:  questionNames,
		QuestionGroups: questionGroupList,
	}
	log.Printf("HandleFormGet(): view: %v\n", view)

	self.renderTemplate(w, "form", view)
}

func (self *App) GetFormUrl(version *entity.Version) (url.URL, error) {
	url := url.URL{
		Path: path.Clean(path.Join(self.FormsRoot, "forms", version.String())),
	}
	return url, nil
}

func (self *App) ParseUrlFormVersion(r *http.Request) (*entity.Version, error) {
	formName := r.URL.Path[len(path.Clean(path.Join(self.FormsRoot, "forms"))+"/"):]
	log.Printf("ParseUrlFormVersion(): formName: %v\n", formName)

	formVer, err := entity.NewVersionFromString(formName)
	checkHTTP(err)
	log.Printf("ParseUrlFormVersion(): formVer: %v\n", formVer)

	return formVer, err
}

func HandleFormPost(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleFormPost()\n")

	formVer, err := self.ParseUrlFormVersion(r)
	checkHTTP(err)
	log.Printf("HandleFormPost(): formVer: %v\n", formVer)

	form, err := self.FormRepo.GetFormById(*formVer)
	checkHTTP(err)
	log.Printf("HandleFormPost(): form: %v\n", form)

	questionName := r.FormValue("question_name")
	log.Printf("HandleFormPost(): questionName: %v\n", questionName)

	questionVer, err := entity.NewVersionFromString(questionName)
	checkHTTP(err)
	log.Printf("HandleFormPost(): questionVer: %v\n", questionVer)

	question := &entity.Question{
		Version: *questionVer,
	}
	oldQuestion, err := self.GetQuestionById(*questionVer)
	if oldQuestion != nil {
		question = oldQuestion
	} else {
		log.Printf("HandleFormPost(): generating fresh question: %v\n", question)
		err = self.QuestionRepo.AddQuestion(question)
		checkHTTP(err)
	}

	form.Questions[*questionVer] = question
	err = self.AddForm(form)
	checkHTTP(err)

	url, err := self.GetFormUrl(formVer)
	checkHTTP(err)
	url.Fragment = "question-" + questionVer.String()

	log.Printf("HandleFormPost(): redirecting to: %v\n", url)
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

	if chartUrl == "/pages" {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleChartSetGet(self, w, r)
		}
		return
	}

	// anyway, assuming it's a chart, find the index.txt
	fi, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			checkHTTP(err)
		}
	}

	if !fi.IsDir() {
		fp3 := fullPath
		// BUG(mistone): don't set Content-Type blindly; also need to check Accept header
		// BUG(mistone): do we really want to sniff mime-types here?
		http.ServeFile(w, r, fp3)
		return
	} else {
		name := path.Join(fullPath, "index.txt")
		if _, err := os.Stat(name); os.IsNotExist(err) {
			name = path.Join(fullPath, "index.text")
			_, err = os.Stat(name)
			checkHTTP(err)
		}

		chart := chart.NewChart(name, self.ChartsPath)

		err = chart.Read()
		checkHTTP(err)

		// attempt to parse header lines
		meta := chart.Meta()

		htmlFlags := 0
		//htmlFlags |= blackfriday.HTML_USE_XHTML
		htmlFlags |= blackfriday.HTML_TOC

		htmlRenderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

		linkRenderer := linker.NewLinkRenderer()

		extFlags := 0
		extFlags |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
		extFlags |= blackfriday.EXTENSION_TABLES
		extFlags |= blackfriday.EXTENSION_FENCED_CODE
		extFlags |= blackfriday.EXTENSION_AUTOLINK
		extFlags |= blackfriday.EXTENSION_STRIKETHROUGH
		extFlags |= blackfriday.EXTENSION_SPACE_HEADERS

		html := blackfriday.Markdown([]byte(chart.Body()), htmlRenderer, extFlags)

		blackfriday.Markdown([]byte(chart.Body()), linkRenderer, extFlags)

		log.Printf("HandleChartGet(): found links: %s", linkRenderer.Links)

		view := &vChart{
			vRoot: newVRoot(self, "chart", meta.Title, meta.Authors, meta.Date),
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
	FormSetUrl     string
	RecordSetUrl   string
}

func newVRoot(self *App, pageName string, title string, authors string, date string) *vRoot {
	questionSetUrl := url.URL{Path: path.Clean(path.Join(self.FormsRoot, "questions"))}
	formSetUrl := url.URL{Path: path.Clean(path.Join(self.FormsRoot, "forms"))}
	recordSetUrl := url.URL{Path: path.Clean(path.Join(self.FormsRoot, "records"))}

	return &vRoot{
		PageName:       pageName,
		Title:          title,
		Authors:        authors,
		Date:           date,
		StaticUrl:      self.StaticRoot,
		FormsRoot:      self.FormsRoot,
		ChartsRoot:     path.Clean(self.ChartsRoot + "/"),
		QuestionSetUrl: questionSetUrl.String(),
		FormSetUrl:     formSetUrl.String(),
		RecordSetUrl:   recordSetUrl.String(),
	}
}

func HandleRootGet(self *App, w http.ResponseWriter, r *http.Request) {
	view := newVRoot(self, "root", "", "", "")
	self.renderTemplate(w, "root", view)
}

func (self *App) GetChartUrl(chart *chart.Chart) (url.URL, error) {
	slug, err := chart.Slug()
	if err != nil {
		return url.URL{}, err
	}
	return url.URL{
		Path: path.Clean(path.Join("/", self.ChartsRoot, slug)) + "/",
	}, nil
}

type vChartLink struct {
	chart.ChartMeta
	Link url.URL
}

type vChartLinkList []*vChartLink

type vChartSet struct {
	*vRoot
	Charts vChartLinkList
}

func HandleChartSetGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleChartSetGet(): start")

	var charts vChartLinkList = nil

	filepath.Walk(self.ChartsPath, func(name string, fi os.FileInfo, err error) error {
		log.Printf("HandleChartSetGet(): visiting path %s", name)
		if err != nil {
			return err
		}

		chart := chart.NewChart(name, self.ChartsPath)

		err = chart.Read()
		if err != nil {
			return nil
		}

		link, err := self.GetChartUrl(chart)

		if err == nil {
			charts = append(charts, &vChartLink{
				ChartMeta: chart.Meta(),
				Link:      link,
			})
		}
		return nil
	})

	now := time.Now()
	date := fmt.Sprintf("%s %0.2d, %d", now.Month().String(), now.Day(), now.Year())

	view := &vChartSet{
		vRoot:  newVRoot(self, "chart_set", "List of Charts", "Michael Stone", date),
		Charts: charts,
	}
	log.Printf("HandleChartSetGet(): view: %s", view)

	self.renderTemplate(w, "chart_set", view)
}

func HandleSiteJsonGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleSiteJsonGet(): start")

	view := map[string]string{}

	filepath.Walk(self.ChartsPath, func(name string, fi os.FileInfo, err error) error {
		//log.Printf("HandleSiteJsonGet(): visiting path %s", name)
		if err != nil {
			return err
		}

		chart := chart.NewChart(name, self.ChartsPath)

		key, err := chart.Slug()
		if err != nil {
			//log.Printf("HandleSiteJsonGet(): warning before read: %s", err)
			return nil
		}
		log.Printf("HandleSiteJsonGet(): found key %s", key)

		err = chart.Read()
		if err != nil {
			log.Printf("HandleSiteJsonGet(): warning after read: %s", err)
			return nil
		}

		bytes := chart.Bytes()
		//log.Printf("HandleSiteJsonGet(): found body: %s", body)

		view[key] = string(bytes)

		return nil
	})

	//log.Printf("HandleSiteJsonGet(): view: %s", view)

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

	if fp == "/records" {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleRecordSetGet(self, w, r)
		case "POST":
			HandleRecordSetPost(self, w, r)
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

	if fp == "/forms" {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleFormSetGet(self, w, r)
		case "POST":
			HandleFormSetPost(self, w, r)
		}
		return
	}

	if strings.HasPrefix(fp, "/records/") {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleRecordGet(self, w, r)
		case "POST":
			HandleRecordPost(self, w, r)
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

	if strings.HasPrefix(fp, "/forms/") {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleFormGet(self, w, r)
		case "POST":
			HandleFormPost(self, w, r)
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
		return
	}

	isForm := strings.HasPrefix(r.URL.Path, path.Clean("/"+self.FormsRoot))
	if isForm {
		log.Printf("HandleRootApp: dispatching to forms...")
		self.HandleForms(w, r)
		return
	}

	isChart := strings.HasPrefix(r.URL.Path, path.Clean("/"+self.ChartsRoot))
	if isChart {
		log.Printf("HandleRootApp: dispatching to charts...")
		HandleChartGet(self, w, r)
		return
	}

	log.Printf("warning: can't route path: %v", r.URL.Path)
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
