// Package persist provides implementations for the domain repos
package persist

import (
	"entity"
	"errors"
	"fmt"
)

type PersistMem struct {
	Profiles  []*entity.Profile
	Questions []*entity.Question
	Reviews   []*entity.Review
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
	profiles, err := self.GetAllProfiles()
	if err != nil {
		return nil, err
	}
	for _, prof := range profiles {
		if prof.Version == version {
			return prof, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("PersistMem.GetProfileById(): profile version '%v' not found", version))
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
	return self.Reviews, nil
}

func (self *PersistMem) GetReviewById(version entity.Version) (*entity.Review, error) {
	reviews, err := self.GetAllReviews()
	if err != nil {
		return nil, err
	}
	for _, rev := range reviews {
		if rev.Version == version {
			return rev, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("PersistMem.GetReviewById(): review version '%v' not found", version))
}

func (self *PersistMem) AddReview(review *entity.Review) error {
	found := false
	for idx, rev := range self.Reviews {
		if rev.Version == review.Version {
			found = true
			self.Reviews[idx] = review
		}
	}
	if !found {
		self.Reviews = append(self.Reviews, review)
	}
	return nil
}
