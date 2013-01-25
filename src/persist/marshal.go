package persist

import (
	"entity"
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

type profileV1 struct {
	Version      versionV1
	QuestionIds  []questionIdV1
	QuestionDeps []questionDepV1
}

type profileSetV1 struct {
	Profiles []profileV1
}

type answerV1 struct {
	Author               string
	CreationTimeSecs     int64
	CreationTimeNanoSecs int64
	Datum                string
}

type reviewV1 struct {
	Version   versionV1
	ProfileId versionV1
	Responses map[string]answerV1
}

type reviewSetV1 struct {
	Reviews []reviewV1
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

func persistProfileSetV1ToEntityProfilePtrSlice(ps profileSetV1, questionsMap map[entity.Version]*entity.Question) []*entity.Profile {
	root := make([]*entity.Profile, len(ps.Profiles))
	for k, v := range ps.Profiles {
		root[k] = &entity.Profile{
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

func entityReviewToPersistReviewV1(r *entity.Review) reviewV1 {
	return reviewV1{
		Version:   entityVersionToPersistVersionV1(r.Version),
		ProfileId: entityVersionToPersistVersionV1(r.Profile.Version),
		Responses: entityResponseMapToPersistResponsesMapV1(r.Responses),
	}
}

func entityReviewPtrSliceToPersistReviewSetV1(rs []*entity.Review) reviewSetV1 {
	root := reviewSetV1{
		Reviews: make([]reviewV1, len(rs)),
	}
	for k, v := range rs {
		root.Reviews[k] = entityReviewToPersistReviewV1(v)
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

func persistResponseMapV1ToEntityResponseMap(profile *entity.Profile, r map[string]answerV1) map[entity.Version]*entity.Response {
	root := make(map[entity.Version]*entity.Response, len(r))
	for k, v := range r {
		questionVer, err := entity.NewVersionFromString(k)
		if err != nil {
			panic(err)
		}
		question, ok := profile.Questions[*questionVer]
		if !ok {
			panic(fmt.Sprintf("unable to find questionVer: %v", questionVer))
		}
		root[*questionVer] = persistResponseV1ToEntityResponse(question, v)
	}
	return root
}

func persistReviewV1ToEntityReview(r reviewV1) *entity.Review {
	var persist *PersistJSON = nil
	profileVer := persistVersionV1ToEntityVersion(r.ProfileId)
	profile, err := persist.jsonGetProfileById(profileVer)
	if err != nil {
		panic(err)
	}
	return &entity.Review{
		Version:   persistVersionV1ToEntityVersion(r.Version),
		Profile:   profile,
		Responses: persistResponseMapV1ToEntityResponseMap(profile, r.Responses),
	}
}

func persistReviewSetV1ToEntityReviewPtrSlice(rs reviewSetV1) []*entity.Review {
	root := make([]*entity.Review, len(rs.Reviews))
	for k, v := range rs.Reviews {
		root[k] = persistReviewV1ToEntityReview(v)
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

func entityProfileToPersistProfileV1(profile *entity.Profile) (profileV1, error) {
	view := profileV1{
		Version:      entityVersionToPersistVersionV1(profile.Version),
		QuestionIds:  make([]questionIdV1, len(profile.Questions)),
		QuestionDeps: make([]questionDepV1, len(profile.QuestionDeps)),
	}
	idx := 0
	for _, question := range profile.Questions {
		ver := entityVersionToPersistVersionV1(question.Version)
		view.QuestionIds[idx] = questionIdV1{
			Version: ver,
		}
		idx++
	}
	// BUG(mistone): question dep persistence not implemented!
	if len(profile.QuestionDeps) > 0 {
		panic("question dep persistence not implemented!")
	}
	return view, nil
}

func entityProfilePtrSliceToPersistProfileSetV1(profiles []*entity.Profile) (profileSetV1, error) {
	view := &profileSetV1{
		Profiles: make([]profileV1, len(profiles)),
	}
	for idx, v := range profiles {
		vprof, err := entityProfileToPersistProfileV1(v)
		if err != nil {
			return *view, err
		} else {
			view.Profiles[idx] = vprof
		}
	}
	return *view, nil
}
