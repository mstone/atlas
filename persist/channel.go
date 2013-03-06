package persist

import (
	"akamai/atlas/forms/entity"
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

type getAllFormsOpRx struct {
	Val []*entity.Form
	Err error
}

type getAllFormsOp struct {
	Persist   *PersistJSON
	ReplyChan chan getAllFormsOpRx
}

type getFormByIdOpRx struct {
	Val *entity.Form
	Err error
}

type getFormByIdOp struct {
	Persist   *PersistJSON
	ReplyChan chan getFormByIdOpRx
	Id        entity.Version
}

type addFormOpRx struct {
	Err error
}

type addFormOp struct {
	Persist   *PersistJSON
	ReplyChan chan addFormOpRx
	Form   *entity.Form
}

type getAllRecordsOpRx struct {
	Val []*entity.Record
	Err error
}

type getAllRecordsOp struct {
	Persist   *PersistJSON
	ReplyChan chan getAllRecordsOpRx
}

type getRecordByIdOpRx struct {
	Val *entity.Record
	Err error
}

type getRecordByIdOp struct {
	Persist   *PersistJSON
	ReplyChan chan getRecordByIdOpRx
	Id        entity.Version
}

type addRecordOpRx struct {
	Err error
}

type addRecordOp struct {
	Persist   *PersistJSON
	ReplyChan chan addRecordOpRx
	Record    *entity.Record
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
		case getAllFormsOp:
			val, err := op.Persist.jsonGetAllForms()
			op.ReplyChan <- getAllFormsOpRx{val, err}
		case getFormByIdOp:
			val, err := op.Persist.jsonGetFormById(op.Id)
			op.ReplyChan <- getFormByIdOpRx{val, err}
		case addFormOp:
			err := op.Persist.jsonAddForm(op.Form)
			op.ReplyChan <- addFormOpRx{err}
		case getAllRecordsOp:
			val, err := op.Persist.jsonGetAllRecords()
			op.ReplyChan <- getAllRecordsOpRx{val, err}
		case getRecordByIdOp:
			val, err := op.Persist.jsonGetRecordById(op.Id)
			op.ReplyChan <- getRecordByIdOpRx{val, err}
		case addRecordOp:
			err := op.Persist.jsonAddRecord(op.Record)
			op.ReplyChan <- addRecordOpRx{err}
		}
	}
}

func dispatchOpChanLoop(opChan opChan) {
	for {
		dispatchOpChan(opChan)
	}
}
