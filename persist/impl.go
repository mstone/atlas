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
	log.Printf("jsonAddProfile(): encoded questionSetV1: %v", view)

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

func (self *PersistJSON) jsonGetAllProfiles() ([]*entity.Profile, error) {
	log.Printf("PersistJSON.GetAllProfiles(): getting all questions...")
	questions, err := self.jsonGetAllQuestions()
	if err != nil {
		return nil, err
	}
	questionsMap := makeQuestionsMap(questions)

	f, err := self.jsonReadPath(self.profilesPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	log.Printf("PersistJSON.GetAllProfiles(): data/profiles.json opened")

	decoder := json.NewDecoder(bufio.NewReader(f))
	log.Printf("PersistJSON.GetAllProfiles(): made decoder")

	ps := profileSetV1{}
	err = decoder.Decode(&ps)
	if err != nil {
		return nil, err
	}
	log.Printf("PersistJSON.GetAllProfiles(): decoded profileSet: %v", ps)

	root := persistProfileSetV1ToEntityProfilePtrSlice(ps, questionsMap)
	return root, nil
}

func (self *PersistJSON) jsonGetProfileById(id entity.Version) (*entity.Profile, error) {
	profiles, err := self.jsonGetAllProfiles()
	if err != nil {
		return nil, err
	}
	for _, prof := range profiles {
		if prof.Version == id {
			return prof, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("PersistJSON.GetProfileById(): profile version '%v' not found", id))
}

func (self *PersistJSON) jsonAddProfile(profile *entity.Profile) error {
	err := self.jsonAddProfileHelper(profile)
	if err == nil {
		err = self.jsonRenameTmpPath(self.profilesPath)
	}
	return err
}

func (self *PersistJSON) jsonAddProfileHelper(profile *entity.Profile) error {
	log.Printf("jsonAddProfile(): adding new profile: %v", profile)
	allProfs, err := self.jsonGetAllProfiles()
	log.Printf("jsonAddProfile(): profs: %v", allProfs)
	if err != nil {
		return err
	}

	found := false
	for idx, prof := range allProfs {
		if prof.Version == profile.Version {
			found = true
			allProfs[idx] = profile
		}
	}
	if !found {
		allProfs = append(allProfs, profile)
	}

	f, err := self.jsonWriteTmpPath(self.profilesPath)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Printf("jsonAddProfile(): data/profiles.json.tmp opened for write")

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	encoder := json.NewEncoder(writer)
	log.Printf("jsonAddProfile(): made encoder")

	view, err := entityProfilePtrSliceToPersistProfileSetV1(allProfs)
	if err != nil {
		return err
	}

	err = encoder.Encode(&view)
	if err != nil {
		return err
	}
	log.Printf("jsonAddProfile(): encoded profileSetV1: %v", view)

	return nil
}

func (self *PersistJSON) jsonGetAllReviews() ([]*entity.Review, error) {
	f, err := self.jsonReadPath(self.reviewsPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	log.Printf("PersistJSON.GetAllReviews(): data/reviews.json opened")

	decoder := json.NewDecoder(bufio.NewReader(f))
	log.Printf("PersistJSON.GetAllReviews(): made decoder")

	rs := reviewSetV1{}
	err = decoder.Decode(&rs)
	if err != nil {
		return nil, err
	}
	log.Printf("PersistJSON.GetAllReviews(): decoded reviewSet: %v", rs)

	return persistReviewSetV1ToEntityReviewPtrSlice(rs), nil
}

func (self *PersistJSON) jsonGetReviewById(id entity.Version) (*entity.Review, error) {
	reviews, err := self.jsonGetAllReviews()
	if err != nil {
		return nil, err
	}
	for _, rev := range reviews {
		if rev.Version == id {
			return rev, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("PersistJSON.GetReviewById(): review version '%v' not found", id))
}

func (self *PersistJSON) jsonAddReview(review *entity.Review) error {
	err := self.jsonAddReviewHelper(review)
	if err == nil {
		err = self.jsonRenameTmpPath(self.reviewsPath)
	}
	return err
}

func (self *PersistJSON) jsonAddReviewHelper(review *entity.Review) error {
	allRevs, err := self.jsonGetAllReviews()
	log.Printf("PersistJSON.AddReview(): revs: %v", allRevs)
	if err != nil {
		return err
	}

	found := false
	for idx, rev := range allRevs {
		if rev.Version == review.Version {
			found = true
			allRevs[idx] = review
		}
	}
	if !found {
		allRevs = append(allRevs, review)
	}

	f, err := self.jsonWriteTmpPath(self.reviewsPath)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Printf("PersistJSON.AddReview(): data/reviews.json.tmp opened for write")

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	encoder := json.NewEncoder(writer)
	log.Printf("PersistJSON.AddReview(): made encoder: %v", encoder)

	rs := entityReviewPtrSliceToPersistReviewSetV1(allRevs)
	log.Printf("PersistJSON.AddReview(): made reviewSetV1: %v", rs)

	err = encoder.Encode(rs)
	if err != nil {
		return err
	}
	log.Printf("PersistJSON.AddReview(): encoded reviewSet: %v", rs)

	return nil
}
