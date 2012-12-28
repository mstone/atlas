// Package entity provides the domain syntax for the review wizard.
// For more information, see 
//
//     http://manuel.kiessling.net/2012/09/28/applying-the-clean-architecture-to-go-applications/
package entity

import (
	"time"
	"strings"
	"fmt"
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
	AddReview(*Review) error
}

func (self *Version) String() string {
	return fmt.Sprintf("%s-%d.%d.%d", self.Name, self.Major, self.Minor, self.Patch)
}

func NewVersionFromString(str string) (*Version, error) {
	cmps := strings.Split(str, "-")
	lhs := strings.Join(cmps[0:len(cmps)-1], "-")
	rhs := cmps[len(cmps)-1]
	var major, minor, patch int
	_, err := fmt.Sscanf(rhs, "%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		return nil, err
	}
	ver := &Version{
		Name: lhs,
		Major: major,
		Minor: minor,
		Patch: patch,
	}
	return ver, nil
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
