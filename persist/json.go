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
	formsPath  string
	recordsPath   string
}

func NewPersistJSON(dataPath string) *PersistJSON {
	log.Printf("PersistJSON.NewPersistJSON(): dataPath: %s", dataPath)

	jsonOpChan := make(chan interface{})

	go dispatchOpChanLoop(jsonOpChan)

	return &PersistJSON{
		opChan:        opChan(jsonOpChan),
		dataPath:      dataPath,
		questionsPath: path.Join(dataPath, "questions.json"),
		formsPath:  path.Join(dataPath, "forms.json"),
		recordsPath:   path.Join(dataPath, "records.json"),
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

func (self *PersistJSON) AddQuestion(form *entity.Question) error {
	replyChan := make(chan addQuestionOpRx)
	op := addQuestionOp{
		Persist:   self,
		ReplyChan: replyChan,
		Question:  form,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Err
}

func (self *PersistJSON) GetAllForms() ([]*entity.Form, error) {
	replyChan := make(chan getAllFormsOpRx)
	op := getAllFormsOp{
		Persist:   self,
		ReplyChan: replyChan,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Val, rx.Err
}

func (self *PersistJSON) GetFormById(version entity.Version) (*entity.Form, error) {
	replyChan := make(chan getFormByIdOpRx)
	op := getFormByIdOp{
		Persist:   self,
		ReplyChan: replyChan,
		Id:        version,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Val, rx.Err
}

func (self *PersistJSON) AddForm(form *entity.Form) error {
	replyChan := make(chan addFormOpRx)
	op := addFormOp{
		Persist:   self,
		ReplyChan: replyChan,
		Form:   form,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Err
}

func (self *PersistJSON) GetAllRecords() ([]*entity.Record, error) {
	replyChan := make(chan getAllRecordsOpRx)
	op := getAllRecordsOp{
		Persist:   self,
		ReplyChan: replyChan,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Val, rx.Err
}

func (self *PersistJSON) GetRecordById(version entity.Version) (*entity.Record, error) {
	replyChan := make(chan getRecordByIdOpRx)
	op := getRecordByIdOp{
		Persist:   self,
		ReplyChan: replyChan,
		Id:        version,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Val, rx.Err
}

func (self *PersistJSON) AddRecord(record *entity.Record) error {
	replyChan := make(chan addRecordOpRx)
	op := addRecordOp{
		Persist:   self,
		ReplyChan: replyChan,
		Record:    record,
	}
	self.opChan <- op
	rx := <-replyChan
	return rx.Err
}
