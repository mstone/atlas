// Package persist provides implementations for the domain repos
package persist

import (
	"errors"
	"entity"
)

type PersistMem struct {
	Profiles  []*entity.Profile
	Questions []*entity.Question
}

func (self *PersistMem) GetAllProfiles() ([]*entity.Profile, error) {
	v := make([]*entity.Profile, 1)
	v[0] = &entity.Profile{
		Version: entity.Version{"pace", 1, 0, 0},
		Questions: []*entity.Question{
			&entity.Question{
				Version:     entity.Version{"who", 1, 0, 0},
				Text:        "Who?",
				Help:        "...'s on first?",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
		},
	}
	return v, nil
}

func (self *PersistMem) GetProfileById(version entity.Version) (*entity.Profile, error) {
	return nil, errors.New("PersistMem.GetProfileById not implemented")
}

func (self *PersistMem) GetAllQuestions() ([]*entity.Question, error) {
	v := make([]*entity.Question, 1)
	v[0] = &entity.Question{
				Version:     entity.Version{"who", 1, 0, 0},
				Text:        "Who?",
				Help:        "...'s on first?",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			}
	return v, nil
}

func (self *PersistMem) GetQuestionById(version entity.Version) (*entity.Question, error) {
	return nil, errors.New("PersistMem.GetQuestionById not implemented")
}

func (self *PersistMem) GetAllReviews() ([]*entity.Review, error) {
	return nil, errors.New("PersistMem.GetAllReviews not implemented")
}

func (self *PersistMem) GetReviewById(version entity.Version) (*entity.Review, error) {
	return nil, errors.New("PersistMem.GetReviewsById not implemented")
}
