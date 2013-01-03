// ARGH
package persist

import (
	//"bufio"
	//"encoding/json"
	"entity"
	"errors"
	//"fmt"
	"os"
	"sync"
)

type PersistJSON struct {
	*sync.RWMutex
}

func NewPersistJSON() (*PersistJSON) {
	return &PersistJSON {
		RWMutex: &sync.RWMutex{},
	}
}

func (self *PersistJSON) GetAllProfiles() ([]*entity.Profile, error) {
	self.RLock()
	defer self.RUnlock()

	f, err := os.Open("data/data.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return nil, errors.New("PersistJSON.GetAllProfiles(): not implemented")
}

func (self *PersistJSON) GetProfileById(version entity.Version) (*entity.Profile, error) {
	self.RLock()
	defer self.RUnlock()

	return nil, errors.New("PersistJSON.GetAllProfiles(): not implemented")
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
