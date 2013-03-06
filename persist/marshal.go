package persist

import (
	"akamai/atlas/forms/entity"
	"fmt"
	"time"
)

type versionV1 struct {
	Name  string
	Major int
	Minor int
	Patch int
}

type questionIdV1 struct {
	Version versionV1
}

type questionV1 struct {
	Version     versionV1
	Text        string
	Help        string
	AnswerType  int
	DisplayHint string
	GroupKey    string
	SortKey     int
}

type questionDepV1 struct{}

type questionSetV1 struct {
	Questions []questionV1
}

type formV1 struct {
	Version      versionV1
	QuestionIds  []questionIdV1
	QuestionDeps []questionDepV1
}

type formSetV1 struct {
	Forms []formV1
}

type answerV1 struct {
	Author               string
	CreationTimeSecs     int64
	CreationTimeNanoSecs int64
	Datum                string
}

type recordV1 struct {
	Version   versionV1
	FormId versionV1
	Responses map[string]answerV1
}

type recordSetV1 struct {
	Records []recordV1
}

func persistVersionV1ToEntityVersion(v versionV1) entity.Version {
	return entity.Version{
		Name:  v.Name,
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
	}
}

func entityVersionToPersistVersionV1(v entity.Version) versionV1 {
	return versionV1{
		Name:  v.Name,
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
	}
}

func persistQuestionV1ToEntityQuestion(q questionV1) *entity.Question {
	return &entity.Question{
		Version:     persistVersionV1ToEntityVersion(q.Version),
		Text:        q.Text,
		Help:        q.Help,
		AnswerType:  q.AnswerType,
		DisplayHint: q.DisplayHint,
		GroupKey:    q.GroupKey,
		SortKey:     q.SortKey,
	}
}

func persistQuestionSetV1ToEntityQuestionPtrSlice(qs questionSetV1) []*entity.Question {
	root := make([]*entity.Question, len(qs.Questions))
	for k, v := range qs.Questions {
		root[k] = persistQuestionV1ToEntityQuestion(v)
	}
	return root
}

func entityQuestionPtrSliceToPersistQuestionSetV1(questions []*entity.Question) (questionSetV1, error) {
	view := &questionSetV1{
		Questions: make([]questionV1, len(questions)),
	}
	for idx, v := range questions {
		vquest, err := entityQuestionToPersistQuestionV1(v)
		if err != nil {
			return *view, err
		} else {
			view.Questions[idx] = vquest
		}
	}
	return *view, nil
}

func persistFormSetV1ToEntityFormPtrSlice(ps formSetV1, questionsMap map[entity.Version]*entity.Question) []*entity.Form {
	root := make([]*entity.Form, len(ps.Forms))
	for k, v := range ps.Forms {
		root[k] = &entity.Form{
			Version:      persistVersionV1ToEntityVersion(v.Version),
			Questions:    make(map[entity.Version]*entity.Question),
			QuestionDeps: make(map[entity.Version]*entity.QuestionDep),
		}
		for _, v2 := range v.QuestionIds {
			questionVer := persistVersionV1ToEntityVersion(v2.Version)
			root[k].Questions[questionVer] = questionsMap[questionVer]
		}
		// BUG(mistone): QuestionDeps not propagated
	}
	return root
}

func entityAnswerToPersistAnswerV1(a *entity.Answer) answerV1 {
	return answerV1{
		Author:               a.Author,
		CreationTimeSecs:     a.CreationTime.Unix(),
		CreationTimeNanoSecs: a.CreationTime.UnixNano(),
		Datum:                a.Datum,
	}
}

func entityResponseMapToPersistResponsesMapV1(rs map[entity.Version]*entity.Response) map[string]answerV1 {
	m := make(map[string]answerV1, len(rs))
	for k, v := range rs {
		m[k.String()] = entityAnswerToPersistAnswerV1(v.Answer)
	}
	return m
}

func entityRecordToPersistRecordV1(r *entity.Record) recordV1 {
	return recordV1{
		Version:   entityVersionToPersistVersionV1(r.Version),
		FormId: entityVersionToPersistVersionV1(r.Form.Version),
		Responses: entityResponseMapToPersistResponsesMapV1(r.Responses),
	}
}

func entityRecordPtrSliceToPersistRecordSetV1(rs []*entity.Record) recordSetV1 {
	root := recordSetV1{
		Records: make([]recordV1, len(rs)),
	}
	for k, v := range rs {
		root.Records[k] = entityRecordToPersistRecordV1(v)
	}
	return root
}

func persistAnswerV1ToEntityAnswer(a answerV1) *entity.Answer {
	return &entity.Answer{
		Author:       a.Author,
		CreationTime: time.Unix(a.CreationTimeSecs, a.CreationTimeNanoSecs),
		Datum:        a.Datum,
	}
}

func persistResponseV1ToEntityResponse(question *entity.Question, ans answerV1) *entity.Response {
	return &entity.Response{
		Question: question,
		Answer:   persistAnswerV1ToEntityAnswer(ans),
	}
}

func persistResponseMapV1ToEntityResponseMap(form *entity.Form, r map[string]answerV1) map[entity.Version]*entity.Response {
	root := make(map[entity.Version]*entity.Response, len(r))
	for k, v := range r {
		questionVer, err := entity.NewVersionFromString(k)
		if err != nil {
			panic(err)
		}
		question, ok := form.Questions[*questionVer]
		if !ok {
			panic(fmt.Sprintf("unable to find questionVer: %v", questionVer))
		}
		root[*questionVer] = persistResponseV1ToEntityResponse(question, v)
	}
	return root
}

func (self *PersistJSON) persistRecordV1ToEntityRecord(r recordV1) *entity.Record {
	formVer := persistVersionV1ToEntityVersion(r.FormId)
	form, err := self.jsonGetFormById(formVer)
	if err != nil {
		panic(err)
	}
	return &entity.Record{
		Version:   persistVersionV1ToEntityVersion(r.Version),
		Form:   form,
		Responses: persistResponseMapV1ToEntityResponseMap(form, r.Responses),
	}
}

func (self *PersistJSON) persistRecordSetV1ToEntityRecordPtrSlice(rs recordSetV1) []*entity.Record {
	root := make([]*entity.Record, len(rs.Records))
	for k, v := range rs.Records {
		root[k] = self.persistRecordV1ToEntityRecord(v)
	}
	return root
}

func entityQuestionToPersistQuestionV1(question *entity.Question) (questionV1, error) {
	view := questionV1{
		Version:     entityVersionToPersistVersionV1(question.Version),
		Text:        question.Text,
		Help:        question.Help,
		AnswerType:  question.AnswerType,
		DisplayHint: question.DisplayHint,
		GroupKey:    question.GroupKey,
		SortKey:     question.SortKey,
	}
	return view, nil
}

func entityFormToPersistFormV1(form *entity.Form) (formV1, error) {
	view := formV1{
		Version:      entityVersionToPersistVersionV1(form.Version),
		QuestionIds:  make([]questionIdV1, len(form.Questions)),
		QuestionDeps: make([]questionDepV1, len(form.QuestionDeps)),
	}
	idx := 0
	for _, question := range form.Questions {
		ver := entityVersionToPersistVersionV1(question.Version)
		view.QuestionIds[idx] = questionIdV1{
			Version: ver,
		}
		idx++
	}
	// BUG(mistone): question dep persistence not implemented!
	if len(form.QuestionDeps) > 0 {
		panic("question dep persistence not implemented!")
	}
	return view, nil
}

func entityFormPtrSliceToPersistFormSetV1(forms []*entity.Form) (formSetV1, error) {
	view := &formSetV1{
		Forms: make([]formV1, len(forms)),
	}
	for idx, v := range forms {
		vprof, err := entityFormToPersistFormV1(v)
		if err != nil {
			return *view, err
		} else {
			view.Forms[idx] = vprof
		}
	}
	return *view, nil
}
