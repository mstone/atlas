// ARGH
//     coupling: the persistence stuff is using entity objects with tight coupling?

package persist

import (
	"bufio"
	"encoding/json"
	"entity"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

type getAllProfilesOpRx struct {
	Val []*entity.Profile
	Err error
}

type getAllProfilesOp struct {
	Persist   *PersistJSON
	ReplyChan chan getAllProfilesOpRx
}

type getProfileByIdOpRx struct {
	Val *entity.Profile
	Err error
}

type getProfileByIdOp struct {
	Persist   *PersistJSON
	ReplyChan chan getProfileByIdOpRx
	Id        entity.Version
}

type addProfileOpRx struct {
	Err error
}

type addProfileOp struct {
	Persist   *PersistJSON
	ReplyChan chan addProfileOpRx
	Profile   *entity.Profile
}

type getAllReviewsOpRx struct {
	Val []*entity.Review
	Err error
}

type getAllReviewsOp struct {
	Persist   *PersistJSON
	ReplyChan chan getAllReviewsOpRx
}

type getReviewByIdOpRx struct {
	Val *entity.Review
	Err error
}

type getReviewByIdOp struct {
	Persist   *PersistJSON
	ReplyChan chan getReviewByIdOpRx
	Id        entity.Version
}

type addReviewOpRx struct {
	Err error
}

type addReviewOp struct {
	Persist   *PersistJSON
	ReplyChan chan addReviewOpRx
	Review    *entity.Review
}

// Messages on opChan will be type-switched to Ops. (See above.)
type opChan chan interface{}

var jsonOpChan = make(chan interface{})

type versionV1 struct {
	Name  string
	Major int
	Minor int
	Patch int
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

type profileV1 struct {
	Version      versionV1
	Questions    []questionV1
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

type PersistJSON struct {
	opChan opChan
}

func NewPersistJSON() *PersistJSON {
	return &PersistJSON{
		opChan: opChan(jsonOpChan),
	}
}

type profileSet struct {
	Profiles []*entity.Profile
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

func persistQuestionV1ToEntityQuestion(q questionV1) entity.Question {
	return entity.Question{
		Version:     persistVersionV1ToEntityVersion(q.Version),
		Text:        q.Text,
		Help:        q.Help,
		AnswerType:  q.AnswerType,
		DisplayHint: q.DisplayHint,
		GroupKey:    q.GroupKey,
		SortKey:     q.SortKey,
	}
}

func persistProfileSetV1ToEntityProfilePtrSlice(ps profileSetV1) []*entity.Profile {
	root := make([]*entity.Profile, len(ps.Profiles))
	for k, v := range ps.Profiles {
		root[k] = &entity.Profile{
			Version:      persistVersionV1ToEntityVersion(v.Version),
			Questions:    make(map[entity.Version]*entity.Question),
			QuestionDeps: make(map[entity.Version]*entity.QuestionDep),
		}
		for _, v2 := range v.Questions {
			questionVer := persistVersionV1ToEntityVersion(v2.Version)
			question := persistQuestionV1ToEntityQuestion(v2)
			root[k].Questions[questionVer] = &question
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

func init() {
	go func() {
		for {
			select {
			case opIface := <-jsonOpChan:
				switch op := opIface.(type) {
				default:
					panic(fmt.Sprintf("persist: unknown operation: %v", op))
				case getAllProfilesOp:
					val, err := op.Persist.jsonGetAllProfiles()
					op.ReplyChan <- getAllProfilesOpRx{val, err}
				case getProfileByIdOp:
					val, err := op.Persist.jsonGetProfileById(op.Id)
					op.ReplyChan <- getProfileByIdOpRx{val, err}
				case addProfileOp:
					err := op.Persist.jsonAddProfile(op.Profile)
					op.ReplyChan <- addProfileOpRx{err}
				case getAllReviewsOp:
					val, err := op.Persist.jsonGetAllReviews()
					op.ReplyChan <- getAllReviewsOpRx{val, err}
				case getReviewByIdOp:
					val, err := op.Persist.jsonGetReviewById(op.Id)
					op.ReplyChan <- getReviewByIdOpRx{val, err}
				case addReviewOp:
					err := op.Persist.jsonAddReview(op.Review)
					op.ReplyChan <- addReviewOpRx{err}
				}
			}
		}
	}()
}

func (self *PersistJSON) GetAllProfiles() ([]*entity.Profile, error) {
	replyChan := make(chan getAllProfilesOpRx)
	op := getAllProfilesOp{
		ReplyChan: replyChan,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Val, rx.Err
}

func (self *PersistJSON) jsonOpenProfilesFile() (*os.File, error) {
	return os.Open("data/profiles.json")
}

func (self *PersistJSON) jsonGetAllProfiles() ([]*entity.Profile, error) {
	f, err := self.jsonOpenProfilesFile()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	log.Printf("PersistJSON.GetAllProfiles(): data/profiles.json opened")

	decoder := json.NewDecoder(bufio.NewReader(f))
	log.Printf("PersistJSON.GetAllProfiles(): made decoder")

	ps := profileSetV1{}
	err = decoder.Decode(&ps)
	if err != nil {
		return nil, err
	}
	log.Printf("PersistJSON.GetAllProfiles(): decoded profileSet: %v", ps)

	root := persistProfileSetV1ToEntityProfilePtrSlice(ps)
	return root, nil
}

func (self *PersistJSON) GetProfileById(version entity.Version) (*entity.Profile, error) {
	replyChan := make(chan getProfileByIdOpRx)
	op := getProfileByIdOp{
		ReplyChan: replyChan,
		Id:        version,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Val, rx.Err
}

func (self *PersistJSON) jsonGetProfileById(id entity.Version) (*entity.Profile, error) {
	profiles, err := self.jsonGetAllProfiles()
	if err != nil {
		return nil, err
	}
	for _, prof := range profiles {
		if prof.Version == id {
			return prof, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("PersistJSON.GetProfileById(): profile version '%v' not found", id))
}

func (self *PersistJSON) AddProfile(profile *entity.Profile) error {
	replyChan := make(chan addProfileOpRx)
	op := addProfileOp{
		ReplyChan: replyChan,
		Profile:   profile,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Err
}

func (self *PersistJSON) jsonAddProfile(profile *entity.Profile) error {
	err := self.jsonAddProfileHelper(profile)
	if err == nil {
		err = os.Rename("data/profiles.json.tmp", "data/profiles.json")
	}
	return err
}

func (self *PersistJSON) jsonAddProfileHelper(profile *entity.Profile) error {
	allProfs, err := self.jsonGetAllProfiles()
	log.Printf("jsonAddProfile(): profs: %v", allProfs)
	if err != nil {
		return err
	}

	found := false
	for idx, prof := range allProfs {
		if prof.Version == profile.Version {
			found = true
			allProfs[idx] = prof
		}
	}
	if !found {
		allProfs = append(allProfs, profile)
	}

	f, err := os.OpenFile("data/profiles.json.tmp",
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Printf("jsonAddProfile(): data/profiles.json.tmp opened for write")

	encoder := json.NewEncoder(bufio.NewWriter(f))
	log.Printf("jsonAddProfile(): made encoder")

	ps := profileSet{
		Profiles: allProfs,
	}
	err = encoder.Encode(&ps)
	if err != nil {
		return err
	}
	log.Printf("jsonAddProfile(): encoded profileSet: %v", ps)

	return nil
}

func (self *PersistJSON) GetAllReviews() ([]*entity.Review, error) {
	replyChan := make(chan getAllReviewsOpRx)
	op := getAllReviewsOp{
		ReplyChan: replyChan,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Val, rx.Err
}

func (self *PersistJSON) jsonGetAllReviews() ([]*entity.Review, error) {
	f, err := os.Open("data/reviews.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	log.Printf("PersistJSON.GetAllReviews(): data/reviews.json opened")

	decoder := json.NewDecoder(bufio.NewReader(f))
	log.Printf("PersistJSON.GetAllReviews(): made decoder")

	rs := reviewSetV1{}
	err = decoder.Decode(&rs)
	if err != nil {
		return nil, err
	}
	log.Printf("PersistJSON.GetAllReviews(): decoded reviewSet: %v", rs)

	return persistReviewSetV1ToEntityReviewPtrSlice(rs), nil
}

func (self *PersistJSON) GetReviewById(version entity.Version) (*entity.Review, error) {
	replyChan := make(chan getReviewByIdOpRx)
	op := getReviewByIdOp{
		ReplyChan: replyChan,
		Id:        version,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Val, rx.Err
}

func (self *PersistJSON) jsonGetReviewById(id entity.Version) (*entity.Review, error) {
	reviews, err := self.jsonGetAllReviews()
	if err != nil {
		return nil, err
	}
	for _, rev := range reviews {
		if rev.Version == id {
			return rev, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("PersistJSON.GetReviewById(): review version '%v' not found", id))
}

func (self *PersistJSON) AddReview(review *entity.Review) error {
	replyChan := make(chan addReviewOpRx)
	op := addReviewOp{
		ReplyChan: replyChan,
		Review:    review,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Err
}

func (self *PersistJSON) jsonAddReview(review *entity.Review) error {
	err := self.jsonAddReviewHelper(review)
	if err == nil {
		err = os.Rename("data/reviews.json.tmp", "data/reviews.json")
	}
	return err
}

func (self *PersistJSON) jsonAddReviewHelper(review *entity.Review) error {
	allRevs, err := self.jsonGetAllReviews()
	log.Printf("PersistJSON.AddReview(): revs: %v", allRevs)
	if err != nil {
		return err
	}

	found := false
	for idx, rev := range allRevs {
		if rev.Version == review.Version {
			found = true
			allRevs[idx] = review
		}
	}
	if !found {
		allRevs = append(allRevs, review)
	}

	f, err := os.OpenFile("data/reviews.json.tmp",
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Printf("PersistJSON.AddReview(): data/reviews.json.tmp opened for write")

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	encoder := json.NewEncoder(writer)
	log.Printf("PersistJSON.AddReview(): made encoder: %v", encoder)

	rs := entityReviewPtrSliceToPersistReviewSetV1(allRevs)
	log.Printf("PersistJSON.AddReview(): made reviewSetV1: %v", rs)

	err = encoder.Encode(rs)
	if err != nil {
		return err
	}
	log.Printf("PersistJSON.AddReview(): encoded reviewSet: %v", rs)

	return nil
}
