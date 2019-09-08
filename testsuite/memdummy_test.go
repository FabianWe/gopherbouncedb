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

type MemdummyTestBinding struct {}

func (b MemdummyTestBinding) BeginInstance() gopherbouncedb.UserStorage {
	return gopherbouncedb.NewMemdummyUserStorage()
}

func (b MemdummyTestBinding) ClosteInstance(s gopherbouncedb.UserStorage) {

}

func TestInitMemdummy(t *testing.T) {
	TestInitSuite(MemdummyTestBinding{}, t)
}

func TestInsertMemdummy(t *testing.T) {
	TestInsertSuite(MemdummyTestBinding{}, true, t)
}

func TestMemdummyLookup(t *testing.T) {
	TestLookupSuite(MemdummyTestBinding{}, true, t)
}

func TestMemdummyUpdate(t *testing.T) {
	TestUpdateUserSuite(MemdummyTestBinding{}, true, t)
}

func TestMemdummyDelete(t *testing.T) {
	TestDeleteUserSuite(MemdummyTestBinding{}, true, t)
}
