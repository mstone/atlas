package persist

import (
	"entity"
	"fmt"
)

type getAllQuestionsOpRx struct {
	Val []*entity.Question
	Err error
}

type getAllQuestionsOp struct {
	Persist   *PersistJSON
	ReplyChan chan getAllQuestionsOpRx
}

type getQuestionByIdOpRx struct {
	Val *entity.Question
	Err error
}

type getQuestionByIdOp struct {
	Persist   *PersistJSON
	ReplyChan chan getQuestionByIdOpRx
	Id        entity.Version
}

type addQuestionOpRx struct {
	Err error
}

type addQuestionOp struct {
	Persist   *PersistJSON
	ReplyChan chan addQuestionOpRx
	Question  *entity.Question
}

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

func dispatchOpChan(opChan opChan) {
	select {
	case opIface := <-opChan:
		switch op := opIface.(type) {
		default:
			panic(fmt.Sprintf("persist: unknown operation: %v", op))
		case getAllQuestionsOp:
			val, err := op.Persist.jsonGetAllQuestions()
			op.ReplyChan <- getAllQuestionsOpRx{val, err}
		case getQuestionByIdOp:
			val, err := op.Persist.jsonGetQuestionById(op.Id)
			op.ReplyChan <- getQuestionByIdOpRx{val, err}
		case addQuestionOp:
			err := op.Persist.jsonAddQuestion(op.Question)
			op.ReplyChan <- addQuestionOpRx{err}
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

func dispatchOpChanLoop(opChan opChan) {
	for {
		dispatchOpChan(opChan)
	}
}
