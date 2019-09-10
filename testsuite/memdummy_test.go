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

package testsuite

import (
	"testing"
	"github.com/FabianWe/gopherbouncedb"
)

type memdummyUserTestBinding struct {}

func (b memdummyUserTestBinding) BeginInstance() gopherbouncedb.UserStorage {
	return gopherbouncedb.NewMemdummyUserStorage()
}

func (b memdummyUserTestBinding) CloseInstance(s gopherbouncedb.UserStorage) {

}

func TestInitMemdummy(t *testing.T) {
	TestInitSuite(memdummyUserTestBinding{}, t)
}

func TestInsertMemdummy(t *testing.T) {
	TestInsertSuite(memdummyUserTestBinding{}, true, t)
}

func TestMemdummyLookup(t *testing.T) {
	TestLookupSuite(memdummyUserTestBinding{}, true, t)
}

func TestMemdummyUpdate(t *testing.T) {
	TestUpdateUserSuite(memdummyUserTestBinding{}, true, t)
}

func TestMemdummyDelete(t *testing.T) {
	TestDeleteUserSuite(memdummyUserTestBinding{}, true, t)
}

type memdummySessionTestBinding struct{}

func (b memdummySessionTestBinding) BeginInstance() gopherbouncedb.SessionStorage {
	return gopherbouncedb.NewMemdummySessionStorage()
}

func (b memdummySessionTestBinding) CloseInstance(s gopherbouncedb.SessionStorage) {

}

func TestInitSessionMemdummy(t *testing.T) {
	TestInitSessionSuite(memdummySessionTestBinding{}, t)
}

func TestInsertSessionMemdummy(t *testing.T) {
	TestSessionInsert(memdummySessionTestBinding{}, t)
}

func TestGetSessionMemdummy(t *testing.T) {
	TestSessionGet(memdummySessionTestBinding{}, t)
}

func TestDeleteSessionMemdummy(t *testing.T) {
	TestSessionDelete(memdummySessionTestBinding{}, t)
}

func TestCleanUpSessionMemdummy(t *testing.T) {
	TestSessionClear(memdummySessionTestBinding{}, t)
}

func TestDeleteForUserMemdummy(t *testing.T) {
	TestSessionDeleteForUser(memdummySessionTestBinding{}, t)
}
