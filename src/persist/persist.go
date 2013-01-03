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
				Version:     entity.Version{"agora-pages", 1, 0, 0},
				Text:        "Which agora pages describe your product?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"bart-releases", 1, 0, 0},
				Text:        "Which BART releases are you using?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"billing-method", 1, 0, 0},
				Text:        "What's your billing method?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"code-lines", 1, 0, 0},
				Text:        "What are your code lines?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"component-names", 1, 0, 0},
				Text:        "What are your components called?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"features", 1, 0, 0},
				Text:        "What features does your product build on?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"bugzilla-tickets", 1, 0, 0},
				Text:        "What CRs does your launch close?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"deployed-networks", 1, 0, 0},
				Text:        "What networks will you be deploying to?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"engineering-schedule-contact", 1, 0, 0},
				Text:        "Who can speak authoritatively about your engineering schedule?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"etherpad-links", 1, 0, 0},
				Text:        "Which etherpad should we use to collaborate on writing up your review?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"implementation-contact", 1, 0, 0},
				Text:        "Who can speak authoritatively about your product's technical implementation details?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"implementation-contact", 1, 0, 0},
				Text:        "Who can speak authoritatively about your product's technical implementation details?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"secarch-review-link", 1, 0, 0},
				Text:        "Which SecArch review covers your product?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"olympus-tickets", 1, 0, 0},
				Text:        "Which Olympus KFs does your product address?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"module-design-link", 1, 0, 0},
				Text:        "Where is your module design document?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"product-contacts", 1, 0, 0},
				Text:        "Who can speak authoritatively about your product's strategy, packaging, etc.",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"program-manager", 1, 0, 0},
				Text:        "Who is program-managing your product?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"secarch-advisor", 1, 0, 0},
				Text:        "Who from SecArch is advising you?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"secarch-reviewer", 1, 0, 0},
				Text:        "Who from SecArch is reviewing your product's security considerations?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"technical-design-contact", 1, 0, 0},
				Text:        "Who can speak authoritatively about your product's technical architecture and design?",
				Help:        "",
				AnswerType:  entity.ANSWER_TYPE_TEXT,
				DisplayHint: "",
			},
			&entity.Question{
				Version:     entity.Version{"jira-tickets", 1, 0, 0},
				Text:        "What Jira tickets does this launch or commit address?",
				Help:        "",
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

func (self *PersistMem) AddProfile(profile *entity.Profile) error {
	found := false
	for idx, prof := range self.Profiles {
		if prof.Version == profile.Version {
			found = true
			self.Profiles[idx] = profile
		}
	}
	if !found {
		self.Profiles = append(self.Profiles, profile)
	}
	return nil
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
