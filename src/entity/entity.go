// Package entity provides the domain syntax for the review wizard.
// For more information, see 
//
//     http://manuel.kiessling.net/2012/09/28/applying-the-clean-architecture-to-go-applications/
package entity

import (
	"time"
)

const (
	ANSWER_TYPE_TEXT int = iota
)

type Version struct {
	Name  string
	Major int
	Minor int
	Patch int
}

type Profile struct {
	Version      Version
	Questions    []*Question
	QuestionDeps []*QuestionDep
}

type Question struct {
	Version     Version
	Text        string
	Help        string
	AnswerType  int
	DisplayHint string // does all this UI stuff belong in the question model?
	GroupKey    string
	SortKey     int
}

type QuestionDep struct {
	From        *Question
	To          []*Question
	Text        string
	DisplayHint string
}

type Answer struct {
	Author       string
	CreationTime time.Time
	Datum        string
	// DescribedTime
}

type Response struct {
	*Question
	*Answer
}

type Review struct {
	Version   Version
	Responses []*Response
}

type ProfileRepo interface {
	GetAllProfiles() ([]*Profile, error)
	GetProfileById(version Version) (*Profile, error)
}

type QuestionRepo interface {
	GetAllQuestions() ([]*Question, error)
	GetQuestionById(version Version) (*Question, error)
}

type ReviewRepo interface {
	GetAllReviews() ([]*Review, error)
	GetReviewById(version Version) (*Review, error)
}

func (profile *Profile) AddQuestion(question *Question) {
	profile.Questions = append(profile.Questions, question)
}

func (profile *Profile) AddQuestionDep(dep *QuestionDep) {
	profile.QuestionDeps = append(profile.QuestionDeps, dep)
}

func (dep *QuestionDep) AddQuestionDep(question *Question) {
	dep.To = append(dep.To, question)
}

func (review *Review) AddResponse(resp *Response) {
	review.Responses = append(review.Responses, resp)
}

// Questions: 
//
//   at the domain level, what are the operations?
//
//   what are the entities and what are the values?
//
//   are there also "repo" interfaces here?
//
//   do the "root" profile and reviews live here?
