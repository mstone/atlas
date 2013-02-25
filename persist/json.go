// ARGH
//     coupling: the persistence stuff is using entity objects with tight coupling?

package persist

import (
	"akamai/atlas/forms/entity"
	"log"
	"path"
)

type PersistJSON struct {
	opChan        opChan
	dataPath      string
	questionsPath string
	profilesPath  string
	reviewsPath   string
}

func NewPersistJSON(dataPath string) *PersistJSON {
	log.Printf("PersistJSON.NewPersistJSON(): dataPath: %s", dataPath)

	jsonOpChan := make(chan interface{})

	go dispatchOpChanLoop(jsonOpChan)

	return &PersistJSON{
		opChan:        opChan(jsonOpChan),
		dataPath:      dataPath,
		questionsPath: path.Join(dataPath, "questions.json"),
		profilesPath:  path.Join(dataPath, "profiles.json"),
		reviewsPath:   path.Join(dataPath, "reviews.json"),
	}
}

func (self *PersistJSON) GetAllQuestions() ([]*entity.Question, error) {
	replyChan := make(chan getAllQuestionsOpRx)
	op := getAllQuestionsOp{
		Persist:   self,
		ReplyChan: replyChan,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Val, rx.Err
}

func (self *PersistJSON) GetQuestionById(version entity.Version) (*entity.Question, error) {
	replyChan := make(chan getQuestionByIdOpRx)
	op := getQuestionByIdOp{
		Persist:   self,
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
		Persist:   self,
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
		Persist:   self,
		ReplyChan: replyChan,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Val, rx.Err
}

func (self *PersistJSON) GetProfileById(version entity.Version) (*entity.Profile, error) {
	replyChan := make(chan getProfileByIdOpRx)
	op := getProfileByIdOp{
		Persist:   self,
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
		Persist:   self,
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
		Persist:   self,
		ReplyChan: replyChan,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Val, rx.Err
}

func (self *PersistJSON) GetReviewById(version entity.Version) (*entity.Review, error) {
	replyChan := make(chan getReviewByIdOpRx)
	op := getReviewByIdOp{
		Persist:   self,
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
		Persist:   self,
		ReplyChan: replyChan,
		Review:    review,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Err
}
