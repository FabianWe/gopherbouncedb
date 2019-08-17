// Copyright 2019 Fabian Wenzelmann
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goherbouncedb

import (
	"sync"
	"fmt"
)

// MemdummyUserStorage is an implementation of UserStorage using an in-memory storage.
// It should never be used in production code, instead it serves as a reference implementation and can be used for
// test cases.
type MemdummyUserStorage struct {
	mutex *sync.RWMutex
	idMapping map[UserID]*UserModel
	nameMapping map[string]*UserModel
	mailMapping map[string]*UserModel
	nextID UserID
}

// NewMemdummyUserStorage returns a new storage without any data.
func NewMemdummyUserStorage() *MemdummyUserStorage {
	return &MemdummyUserStorage{
		mutex: new(sync.RWMutex),
		idMapping: make(map[UserID]*UserModel),
		nextID: 0,
	}
}

func (s *MemdummyUserStorage) Init() error {
	return nil
}

func (s *MemdummyUserStorage) GetUser(id UserID) (*UserModel, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	user, has := s.idMapping[id]
	if !has {
		return nil, NewNoSuchUser(fmt.Sprintf("User with id %d does not exist", id))
	}
	return user, nil
}

func (s *MemdummyUserStorage) GetUserByName(username string) (*UserModel, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	user, has := s.nameMapping[username]
	if !has {
		return nil, NewNoSuchUser(fmt.Sprintf("User with username %s does not exist", username))
	}
	return user, nil
}

func (s *MemdummyUserStorage) GetUserByEmail(email string) (*UserModel, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	user, has := s.mailMapping[email]
	if !has {
		return nil, NewNoSuchUser(fmt.Sprintf("User with email %s does not exist",email))
	}
	return user, nil
}

func (s *MemdummyUserStorage) InsertUser(user *UserModel) (UserID, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// check if username or email already in use
	if _, hasName := s.nameMapping[user.Username]; hasName {
		return InvalidUserID, NewUserExists(fmt.Sprintf("User with name %s already exists", user.Username))
	}
	if _, hasMail := s.mailMapping[user.EMail]; hasMail {
		return InvalidUserID, NewUserExists(fmt.Sprintf("User with email %s already exists", user.EMail))
	}
	// get next id
	nextID := s.nextID
	s.nextID++
	user.ID = nextID
	// add to mappings
	s.idMapping[nextID] = user
	s.nameMapping[user.Username] = user
	s.mailMapping[user.Username] = user
	return nextID, nil
}

func (s *MemdummyUserStorage) UpdateUser(id UserID, newCredentials *UserModel, fields []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// fields is ignored, we just update
	// first find the user with the given id
	existing, has := s.idMapping[id]
	if !has {
		return NewNoSuchUser(fmt.Sprintf("User with id %d does not exist", id))
	}
	// next check if the new username is already in use. if yes: update is only allowed if it refers to the same
	// user (this means the username has not changed). Otherwise the username is used by another account and
	// can't be changed

	if fromName, hasName := s.nameMapping[newCredentials.Username]; hasName && fromName.ID != existing.ID {
		return NewAmbiguousCredentials(fmt.Sprintf("Username %s is already in use", newCredentials.Username))
	}
	// same for mail
	if fromMail, hasMail := s.mailMapping[newCredentials.EMail]; hasMail && fromMail.ID != existing.ID {
		return NewAmbiguousCredentials(fmt.Sprintf("User with email %s already exists", newCredentials.EMail))
	}
	// now everything is okay so we just update
	s.idMapping[id] = newCredentials
	// delete entries for username and email, they might have changed
	delete(s.nameMapping, newCredentials.Username)
	delete(s.mailMapping, newCredentials.EMail)
	// set new values
	s.nameMapping[newCredentials.Username] = newCredentials
	s.mailMapping[newCredentials.EMail] = newCredentials
	return nil
}

func (s *MemdummyUserStorage) DeleteUser(id UserID) error {
	// first get username and email, these entries should be deleted as well
	existing, has := s.idMapping[id]
	if !has {
		return NewNoSuchUser(fmt.Sprintf("User with id %d does not exist", id))
	}
	delete(s.nameMapping, existing.Username)
	delete(s.mailMapping, existing.EMail)
	// delete id
	delete(s.idMapping, id)
	return nil
}
