// ARGH
//     coupling: the persistence stuff is using entity objects with tight coupling?

package persist

import (
	"entity"
)

type PersistJSON struct {
	opChan opChan
}

func NewPersistJSON() *PersistJSON {
	jsonOpChan := make(chan interface{})

	go dispatchOpChanLoop(jsonOpChan)

	return &PersistJSON{
		opChan: opChan(jsonOpChan),
	}
}

func (self *PersistJSON) GetAllQuestions() ([]*entity.Question, error) {
	replyChan := make(chan getAllQuestionsOpRx)
	op := getAllQuestionsOp{
		ReplyChan: replyChan,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Val, rx.Err
}

func (self *PersistJSON) GetQuestionById(version entity.Version) (*entity.Question, error) {
	replyChan := make(chan getQuestionByIdOpRx)
	op := getQuestionByIdOp{
		ReplyChan: replyChan,
		Id:        version,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Val, rx.Err
}

func (self *PersistJSON) AddQuestion(profile *entity.Question) error {
	replyChan := make(chan addQuestionOpRx)
	op := addQuestionOp{
		ReplyChan: replyChan,
		Question:  profile,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Err
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

func (self *PersistJSON) GetAllReviews() ([]*entity.Review, error) {
	replyChan := make(chan getAllReviewsOpRx)
	op := getAllReviewsOp{
		ReplyChan: replyChan,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Val, rx.Err
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
