package persist

import (
	"encoding/json"
	"entity"
	"errors"
	"fmt"
	"os"
	"log"
	"bufio"
)

func (self *PersistJSON) jsonGetAllQuestions() ([]*entity.Question, error) {
	return nil, errors.New("jsonGetAllQuestions() not implemented")
}

func (self *PersistJSON) jsonGetQuestionById(id entity.Version) (*entity.Question, error) {
	return nil, errors.New("jsonGetQuestionById() not implemented")
}

func (self *PersistJSON) jsonAddQuestion(*entity.Question) error {
	return errors.New("jsonAddQuestion() not implemented")
}

func (self *PersistJSON) jsonOpenProfilesFile() (*os.File, error) {
	return os.Open("data/profiles.json")
}

func (self *PersistJSON) jsonGetAllProfiles() ([]*entity.Profile, error) {
	f, err := self.jsonOpenProfilesFile()
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

	root := persistProfileSetV1ToEntityProfilePtrSlice(ps)
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
		err = os.Rename("data/profiles.json.tmp", "data/profiles.json")
	}
	return err
}

func (self *PersistJSON) jsonAddProfileHelper(profile *entity.Profile) error {
	allProfs, err := self.jsonGetAllProfiles()
	log.Printf("jsonAddProfile(): profs: %v", allProfs)
	if err != nil {
		return err
	}

	found := false
	for idx, prof := range allProfs {
		if prof.Version == profile.Version {
			found = true
			allProfs[idx] = prof
		}
	}
	if !found {
		allProfs = append(allProfs, profile)
	}

	f, err := os.OpenFile("data/profiles.json.tmp",
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Printf("jsonAddProfile(): data/profiles.json.tmp opened for write")

	encoder := json.NewEncoder(bufio.NewWriter(f))
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
	f, err := os.Open("data/reviews.json")
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
		err = os.Rename("data/reviews.json.tmp", "data/reviews.json")
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

	f, err := os.OpenFile("data/reviews.json.tmp",
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600)
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
