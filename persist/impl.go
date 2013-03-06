package persist

import (
	"akamai/atlas/forms/entity"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

func (self *PersistJSON) jsonGetAllQuestions() ([]*entity.Question, error) {
	log.Printf("PersistJSON.GetAllQuestions()...")
	log.Printf("PersistJSON.GetAllQuestions(): reading questionsPath: %s", self.questionsPath)
	f, err := self.jsonReadPath(self.questionsPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	log.Printf("PersistJSON.GetAllQuestions(): data/questions.json opened")

	decoder := json.NewDecoder(bufio.NewReader(f))
	log.Printf("PersistJSON.GetAllQuestions(): made decoder")

	qs := questionSetV1{}
	err = decoder.Decode(&qs)
	if err != nil {
		return nil, err
	}
	log.Printf("PersistJSON.GetAllQuestions(): decoded questionSet: %v", qs)

	root := persistQuestionSetV1ToEntityQuestionPtrSlice(qs)
	return root, nil
}

func (self *PersistJSON) jsonGetQuestionById(id entity.Version) (*entity.Question, error) {
	questions, err := self.jsonGetAllQuestions()
	if err != nil {
		return nil, err
	}
	for _, quest := range questions {
		if quest.Version == id {
			return quest, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("PersistJSON.GetQuestionById(): question version '%v' not found", id))
}

func (self *PersistJSON) jsonAddQuestion(question *entity.Question) error {
	err := self.jsonAddQuestionHelper(question)
	if err == nil {
		err = self.jsonRenameTmpPath(self.questionsPath)
	}
	return err
}

func (self *PersistJSON) jsonAddQuestionHelper(question *entity.Question) error {
	allQuestions, err := self.jsonGetAllQuestions()
	log.Printf("jsonAddQuestionHelper(): questions: %v", allQuestions)
	if err != nil {
		return err
	}

	found := false
	for idx, quest := range allQuestions {
		if quest.Version == question.Version {
			found = true
			allQuestions[idx] = question
		}
	}
	if !found {
		allQuestions = append(allQuestions, question)
	}

	f, err := self.jsonWriteTmpPath(self.questionsPath)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Printf("jsonAddQuestionHelper(): " + self.questionsPath + ".tmp opened for write")

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	encoder := json.NewEncoder(writer)
	log.Printf("jsonAddQuestionHelper(): made encoder")

	view, err := entityQuestionPtrSliceToPersistQuestionSetV1(allQuestions)
	if err != nil {
		return err
	}

	err = encoder.Encode(&view)
	if err != nil {
		return err
	}
	log.Printf("jsonAddForm(): encoded questionSetV1: %v", view)

	return nil
}

func (self *PersistJSON) jsonReadPath(path string) (*os.File, error) {
	log.Printf("PersistJSON.ReadPath(): reading %s", path)
	return os.Open(path)
}

func (self *PersistJSON) jsonWriteTmpPath(path string) (*os.File, error) {
	log.Printf("PersistJSON.WriteTmpPath(): writing %s.tmp", path)
	return os.OpenFile(path+".tmp",
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600)
}

func (self *PersistJSON) jsonRenameTmpPath(path string) error {
	log.Printf("PersistJSON.RenameTmpPath(): renaming %s{.tmp,}", path)
	return os.Rename(path+".tmp", path)
}

func makeQuestionsMap(questions []*entity.Question) map[entity.Version]*entity.Question {
	questionsMap := make(map[entity.Version]*entity.Question, len(questions))
	for _, v := range questions {
		questionsMap[v.Version] = v
	}
	return questionsMap
}

func (self *PersistJSON) jsonGetAllForms() ([]*entity.Form, error) {
	log.Printf("PersistJSON.GetAllForms(): getting all questions...")
	questions, err := self.jsonGetAllQuestions()
	if err != nil {
		return nil, err
	}
	questionsMap := makeQuestionsMap(questions)

	f, err := self.jsonReadPath(self.formsPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	log.Printf("PersistJSON.GetAllForms(): data/forms.json opened")

	decoder := json.NewDecoder(bufio.NewReader(f))
	log.Printf("PersistJSON.GetAllForms(): made decoder")

	ps := formSetV1{}
	err = decoder.Decode(&ps)
	if err != nil {
		return nil, err
	}
	log.Printf("PersistJSON.GetAllForms(): decoded formSet: %v", ps)

	root := persistFormSetV1ToEntityFormPtrSlice(ps, questionsMap)
	return root, nil
}

func (self *PersistJSON) jsonGetFormById(id entity.Version) (*entity.Form, error) {
	forms, err := self.jsonGetAllForms()
	if err != nil {
		return nil, err
	}
	for _, prof := range forms {
		if prof.Version == id {
			return prof, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("PersistJSON.GetFormById(): form version '%v' not found", id))
}

func (self *PersistJSON) jsonAddForm(form *entity.Form) error {
	err := self.jsonAddFormHelper(form)
	if err == nil {
		err = self.jsonRenameTmpPath(self.formsPath)
	}
	return err
}

func (self *PersistJSON) jsonAddFormHelper(form *entity.Form) error {
	log.Printf("jsonAddForm(): adding new form: %v", form)
	allProfs, err := self.jsonGetAllForms()
	log.Printf("jsonAddForm(): profs: %v", allProfs)
	if err != nil {
		return err
	}

	found := false
	for idx, prof := range allProfs {
		if prof.Version == form.Version {
			found = true
			allProfs[idx] = form
		}
	}
	if !found {
		allProfs = append(allProfs, form)
	}

	f, err := self.jsonWriteTmpPath(self.formsPath)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Printf("jsonAddForm(): data/forms.json.tmp opened for write")

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	encoder := json.NewEncoder(writer)
	log.Printf("jsonAddForm(): made encoder")

	view, err := entityFormPtrSliceToPersistFormSetV1(allProfs)
	if err != nil {
		return err
	}

	err = encoder.Encode(&view)
	if err != nil {
		return err
	}
	log.Printf("jsonAddForm(): encoded formSetV1: %v", view)

	return nil
}

func (self *PersistJSON) jsonGetAllRecords() ([]*entity.Record, error) {
	f, err := self.jsonReadPath(self.recordsPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	log.Printf("PersistJSON.GetAllRecords(): data/records.json opened")

	decoder := json.NewDecoder(bufio.NewReader(f))
	log.Printf("PersistJSON.GetAllRecords(): made decoder")

	rs := recordSetV1{}
	err = decoder.Decode(&rs)
	if err != nil {
		return nil, err
	}
	log.Printf("PersistJSON.GetAllRecords(): decoded recordSet: %v", rs)

	return self.persistRecordSetV1ToEntityRecordPtrSlice(rs), nil
}

func (self *PersistJSON) jsonGetRecordById(id entity.Version) (*entity.Record, error) {
	records, err := self.jsonGetAllRecords()
	if err != nil {
		return nil, err
	}
	for _, rev := range records {
		if rev.Version == id {
			return rev, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("PersistJSON.GetRecordById(): record version '%v' not found", id))
}

func (self *PersistJSON) jsonAddRecord(record *entity.Record) error {
	err := self.jsonAddRecordHelper(record)
	if err == nil {
		err = self.jsonRenameTmpPath(self.recordsPath)
	}
	return err
}

func (self *PersistJSON) jsonAddRecordHelper(record *entity.Record) error {
	allRevs, err := self.jsonGetAllRecords()
	log.Printf("PersistJSON.AddRecord(): revs: %v", allRevs)
	if err != nil {
		return err
	}

	found := false
	for idx, rev := range allRevs {
		if rev.Version == record.Version {
			found = true
			allRevs[idx] = record
		}
	}
	if !found {
		allRevs = append(allRevs, record)
	}

	f, err := self.jsonWriteTmpPath(self.recordsPath)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Printf("PersistJSON.AddRecord(): data/records.json.tmp opened for write")

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	encoder := json.NewEncoder(writer)
	log.Printf("PersistJSON.AddRecord(): made encoder: %v", encoder)

	rs := entityRecordPtrSliceToPersistRecordSetV1(allRevs)
	log.Printf("PersistJSON.AddRecord(): made recordSetV1: %v", rs)

	err = encoder.Encode(rs)
	if err != nil {
		return err
	}
	log.Printf("PersistJSON.AddRecord(): encoded recordSet: %v", rs)

	return nil
}
