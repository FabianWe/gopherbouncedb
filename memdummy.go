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

package gopherbouncedb

import (
	"fmt"
	"sync"
	"time"
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
		nameMapping: make(map[string]*UserModel),
		mailMapping: make(map[string]*UserModel),
		nextID: 1,
	}
}

func (s *MemdummyUserStorage) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.idMapping = make(map[UserID]*UserModel)
	s.nameMapping = make(map[string]*UserModel)
	s.mailMapping = make(map[string]*UserModel)
	s.nextID = 1
}

func (s *MemdummyUserStorage) InitUsers() error {
	return nil
}

func (s *MemdummyUserStorage) GetUser(id UserID) (*UserModel, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	user, has := s.idMapping[id]
	if !has {
		return nil, NewNoSuchUser(fmt.Sprintf("user with id %d does not exist", id))
	}
	return user, nil
}

func (s *MemdummyUserStorage) GetUserByName(username string) (*UserModel, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	user, has := s.nameMapping[username]
	if !has {
		return nil, NewNoSuchUser(fmt.Sprintf("user with username %s does not exist", username))
	}
	return user, nil
}

func (s *MemdummyUserStorage) GetUserByEmail(email string) (*UserModel, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	user, has := s.mailMapping[email]
	if !has {
		return nil, NewNoSuchUser(fmt.Sprintf("user with email %s does not exist",email))
	}
	return user, nil
}

func (s *MemdummyUserStorage) InsertUser(user *UserModel) (UserID, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// check if username or email already in use
	if _, hasName := s.nameMapping[user.Username]; hasName {
		return InvalidUserID, NewUserExists(fmt.Sprintf("user with name %s already exists", user.Username))
	}
	if _, hasMail := s.mailMapping[user.EMail]; hasMail {
		return InvalidUserID, NewUserExists(fmt.Sprintf("user with email %s already exists", user.EMail))
	}
	// get next id
	nextID := s.nextID
	s.nextID++
	user.ID = nextID
	user.DateJoined = time.Now().UTC()
	// add to mappings
	s.idMapping[nextID] = user.Copy()
	s.nameMapping[user.Username] = user.Copy()
	s.mailMapping[user.EMail] = user.Copy()
	return nextID, nil
}

func (s *MemdummyUserStorage) UpdateUser(id UserID, newCredentials *UserModel, fields []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// fields is ignored, we just update
	// first find the user with the given id
	existing, has := s.idMapping[id]
	if !has {
		return NewNoSuchUser(fmt.Sprintf("user with id %d does not exist", id))
	}
	// next check if the new username is already in use. if yes: update is only allowed if it refers to the same
	// user (this means the username has not changed). Otherwise the username is used by another account and
	// can't be changed

	if fromName, hasName := s.nameMapping[newCredentials.Username]; hasName && fromName.ID != existing.ID {
		return NewAmbiguousCredentials(fmt.Sprintf("username %s is already in use", newCredentials.Username))
	}
	// same for mail
	if fromMail, hasMail := s.mailMapping[newCredentials.EMail]; hasMail && fromMail.ID != existing.ID {
		return NewAmbiguousCredentials(fmt.Sprintf("user with email %s already exists", newCredentials.EMail))
	}
	// now everything is okay so we just update
	s.idMapping[id] = newCredentials.Copy()
	// delete entries for username and email, they might have changed
	delete(s.nameMapping, newCredentials.Username)
	delete(s.mailMapping, newCredentials.EMail)
	// set new values
	s.nameMapping[newCredentials.Username] = newCredentials.Copy()
	s.mailMapping[newCredentials.EMail] = newCredentials.Copy()
	return nil
}

func (s *MemdummyUserStorage) DeleteUser(id UserID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	existing, has := s.idMapping[id]
	if !has {
		return nil
	}
	delete(s.nameMapping, existing.Username)
	delete(s.mailMapping, existing.EMail)
	delete(s.idMapping, id)
	return nil
}

type memUserIterator struct {
	// not very efficient, better ways are possible: just store all users in a slice
	items []*UserModel
	pos int
}

func (it *memUserIterator) HasNext() bool {
	return it.pos < len(it.items)
}

func (it *memUserIterator) Next() (*UserModel, error) {
	if !it.HasNext() {
		return nil, nil
	}
	res := it.items[it.pos]
	it.pos++
	return res, nil
}

func (it *memUserIterator) Err() error {
	return nil
}

func (it *memUserIterator) Close() error {
	return nil
}

func newMemUserIterator(s *MemdummyUserStorage) *memUserIterator {
	items := make([]*UserModel, len(s.idMapping))
	i := 0
	for _, u := range s.idMapping {
		items[i] = u.Copy()
		i++
	}
	return &memUserIterator{
		items: items,
		pos:   0,
	}
}

func (s *MemdummyUserStorage) ListUsers() (UserIterator, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return newMemUserIterator(s), nil
}

type MemdummySessionStorage struct {
	mutex *sync.RWMutex
	keyMapping map[string]*SessionEntry
}

func NewMemdummySessionStorage() *MemdummySessionStorage {
	return &MemdummySessionStorage{
		mutex: new(sync.RWMutex),
		keyMapping: make(map[string]*SessionEntry),
	}
}

func (s *MemdummySessionStorage) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.keyMapping = make(map[string]*SessionEntry)
}

func (s *MemdummySessionStorage) InitSessions() error {
	return nil
}

func (s *MemdummySessionStorage) InsertSession(session *SessionEntry) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, exists := s.keyMapping[session.Key]; exists {
		return NewSessionExistsKey(session.Key)
	}
	s.keyMapping[session.Key] = session.Copy()
	return nil
}

func (s *MemdummySessionStorage) GetSession(key string) (*SessionEntry, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if entry, has := s.keyMapping[key]; has {
		return entry, nil
	}
	return nil, NewNoSuchSessionKey(key)
}

func (s *MemdummySessionStorage) DeleteSession(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.keyMapping, key)
	return nil
}

func (s *MemdummySessionStorage) CleanUp(referenceDate time.Time) (int64, error) {
	var delCount int64
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for key, session := range s.keyMapping {
		if !session.IsValid(referenceDate) {
			delete(s.keyMapping, key)
			delCount++
		}
	}
	return delCount, nil
}

func (s *MemdummySessionStorage) DeleteForUser(user UserID) (int64, error) {
	var delCount int64
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for key, session := range s.keyMapping {
		if session.User == user {
			delete(s.keyMapping, key)
			delCount++
		}
	}
	return delCount, nil
}
