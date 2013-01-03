package persist

import (
	"bufio"
	"encoding/json"
	"entity"
	"errors"
	//"fmt"
	"log"
	"os"
	"sync"
)

type PersistJSON struct {
	*sync.RWMutex
}

func NewPersistJSON() *PersistJSON {
	return &PersistJSON{
		RWMutex: &sync.RWMutex{},
	}
}

type profileSet struct {
	Profiles []*entity.Profile
}

func (self *PersistJSON) GetAllProfiles() ([]*entity.Profile, error) {
	self.RLock()
	defer self.RUnlock()

	f, err := os.Open("data/data.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	log.Printf("PersistJSON.GetAllProfiles(): data/data.json opened")

	decoder := json.NewDecoder(bufio.NewReader(f))
	log.Printf("PersistJSON.GetAllProfiles(): made decoder")

	ps := profileSet{}
	err = decoder.Decode(&ps)
	if err != nil {
		return nil, err
	}
	log.Printf("PersistJSON.GetAllProfiles(): decoded ProfileSet: %v", ps)

	return ps.Profiles, nil
}

func (self *PersistJSON) GetProfileById(version entity.Version) (*entity.Profile, error) {
	self.RLock()
	defer self.RUnlock()

	return nil, errors.New("PersistJSON.GetAllProfiles(): not implemented")
}

func (self *PersistJSON) AddProfile(profile *entity.Profile) error {
	profs, err := self.GetAllProfiles()
	log.Printf("PersistJSON.AddProfile(): profs: %v", profs)
	if err != nil {
		return err
	}

	// BUG(mistone): dedup!
	profs = append(profs, profile)

	// BUG(mistone): race!
	self.Lock()
	defer self.Unlock()

	f, err := os.OpenFile("data/data.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Printf("PersistJSON.AddProfile(): data/data.json opened for write")

	encoder := json.NewEncoder(bufio.NewWriter(f))
	log.Printf("PersistJSON.AddProfile(): made encoder")

	ps := profileSet{
		Profiles: profs,
	}
	err = encoder.Encode(&ps)
	if err != nil {
		return err
	}
	log.Printf("PersistJSON.AddProfile(): encoded ProfileSet: %v", ps)

	return nil
}

func (self *PersistJSON) GetAllReviews() ([]*entity.Review, error) {
	self.RLock()
	defer self.RUnlock()

	return nil, errors.New("PersistJSON.GetAllProfiles(): not implemented")
}

func (self *PersistJSON) GetReviewById(version entity.Version) (*entity.Review, error) {
	self.RLock()
	defer self.RUnlock()

	return nil, errors.New("PersistJSON.GetAllProfiles(): not implemented")
}

func (self *PersistJSON) AddReview(review *entity.Review) error {
	self.Lock()
	defer self.Unlock()

	return errors.New("PersistJSON.GetAllProfiles(): not implemented")
}
