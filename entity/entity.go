// Package entity provides the domain syntax for the record wizard.
// The key ideas are:
//
// 1. A form is a set of (grouped, versioned) questions.
//
// 2. Records collect responses to questions in a form.
//
// 3. A dependency graph connects questions. The graph is used to render UI warnings.
package entity

import (
	"errors"
	"fmt"
	"strings"
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

type Form struct {
	Version      Version
	Questions    map[Version]*Question
	QuestionDeps map[Version]*QuestionDep
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
	To          map[Version]*Question
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

type Record struct {
	Version   Version
	Form   *Form
	Responses map[Version]*Response
}

type QuestionRepo interface {
	GetAllQuestions() ([]*Question, error)
	GetQuestionById(version Version) (*Question, error)
	AddQuestion(*Question) error
}

type FormRepo interface {
	GetAllForms() ([]*Form, error)
	GetFormById(version Version) (*Form, error)
	AddForm(*Form) error
}

type RecordRepo interface {
	GetAllRecords() ([]*Record, error)
	GetRecordById(version Version) (*Record, error)
	AddRecord(*Record) error
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
		Name:  lhs,
		Major: major,
		Minor: minor,
		Patch: patch,
	}
	return ver, nil
}

func (self *Record) SetResponseAnswer(questionVer Version, answer *Answer) (*Record, error) {
	err := errors.New(fmt.Sprintf("Record.SetResponseAnswer(): "+
		"question not found; questionVer: %v, "+
		"answer: %v", questionVer, answer))

	question := self.Form.Questions[questionVer]

	if question != nil {
		self.Responses[questionVer] = &Response{
			Question: question,
			Answer:   answer,
		}
		err = nil
	}

	return self, err
}

func NewAnswer() *Answer {
	return &Answer{
		Author:       "",
		CreationTime: time.Unix(0, 0),
		Datum:        "",
	}
}

// Questions: 
//
//   at the domain level, what are the operations?
//
//   what are the entities and what are the values?
//
//   are there also "repo" interfaces here?
//
//   do the "root" form and records live here?
