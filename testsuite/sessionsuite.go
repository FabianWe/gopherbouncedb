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
	"github.com/FabianWe/gopherbouncedb"
	"testing"
	"strings"
	"time"
	"log"
	"fmt"
)

type SessionTestSuiteBinding interface {
	BeginInstance() gopherbouncedb.SessionStorage
	CloseInstance(s gopherbouncedb.SessionStorage)
}

var (
	sessions = make([]*gopherbouncedb.SessionEntry, 0)
)

func init() {
	restoreDefaultsSession()
}

func genKey(t *testing.T) string {
	key, keyErr := gopherbouncedb.GenSessionKey()
	if keyErr != nil {
		t.Fatal("Unable to create key:", keyErr.Error())
	}
	return key
}

func parseTime(s string) time.Time {
	t, parseErr :=time.Parse("02-01-2006", s)
	if parseErr != nil {
		log.Fatal("Can't parse predefined time:", parseErr.Error())
	}
	return t
}

func restoreDefaultsSession() {
	// use fixed keys to make testing consistent
	key1 := strings.Repeat("A", 39)
	key2 := strings.Repeat("B", 39)
	key3 := strings.Repeat("C", 39)
	s1 := &gopherbouncedb.SessionEntry{
		User: 1,
		Key: key1,
		ExpireDate: parseTime("09-09-2019"),
	}
	s2 := &gopherbouncedb.SessionEntry{
		User: 2,
		Key: key2,
		ExpireDate: parseTime("10-09-2019"),
	}
	s3 := &gopherbouncedb.SessionEntry{
		User: 3,
		Key: key3,
		ExpireDate: parseTime("12-09-2019"),
	}
	sessions = []*gopherbouncedb.SessionEntry{s1, s2, s3}
}

func getSessionsOkay() []*gopherbouncedb.SessionEntry {
	return sessions[:3]
}

func TestInitSessionSuite(suite SessionTestSuiteBinding, t *testing.T) {
	restoreDefaultsSession()
	inst := suite.BeginInstance()
	defer suite.CloseInstance(inst)
	initErr := inst.InitSessions()
	if initErr != nil {
		t.Error("Init failed:", initErr)
	}
}

func insertSessionsOkay(inst gopherbouncedb.SessionStorage, t *testing.T) {
	for _, s := range getSessionsOkay() {
		if insertErr := inst.InsertSession(s); insertErr != nil {
			t.Fatal("Unable to insert session:", insertErr.Error())
		}
	}
}

func TestSessionInsert(suite SessionTestSuiteBinding, t *testing.T) {
	restoreDefaultsSession()
	inst := suite.BeginInstance()
	defer suite.CloseInstance(inst)
	initErr := inst.InitSessions()
	if initErr != nil {
		t.Fatal("Init failed:", initErr)
	}
	insertSessionsOkay(inst, t)
	// test that multiple inserts yield an error
	for _, s := range getSessionsOkay() {
		insertErr := inst.InsertSession(s)
		if insertErr == nil {
			t.Error("Insert of duplicate session key succeeded, expected duplicate error")
		} else {
			if _, isSessionExists := insertErr.(gopherbouncedb.SessionExists); !isSessionExists {
				t.Error("Insert of duplicate session returned an unkown error, expected duplicate key error:",
					insertErr.Error())
			}
		}

	}
}

func compareSessions(s1, s2 *gopherbouncedb.SessionEntry) bool {
	return s1.User == s2.User && s1.Key == s2.Key && compareTime(s1.ExpireDate, s2.ExpireDate)
}

func TestSessionGet(suite SessionTestSuiteBinding, t *testing.T) {
	restoreDefaultsSession()
	inst := suite.BeginInstance()
	defer suite.CloseInstance(inst)
	initErr := inst.InitSessions()
	if initErr != nil {
		t.Fatal("Init failed:", initErr)
	}
	insertSessionsOkay(inst, t)
	// now get all three elements
	for _, s := range getSessionsOkay() {
		lookup, lookupErr := inst.GetSession(s.Key)
		if lookupErr != nil {
			t.Error(fmt.Sprintf("Get of key %s returned an error: %s",
				s.Key, lookupErr.Error()))
		}
		if !compareSessions(s, lookup) {
			t.Error(fmt.Sprintf("Get session returned wrong element. Expected %v and got %v",
				s, lookup))
		}
	}
	// test non-existing sessions
	invalidKeys := []string{
		strings.Repeat("X", 39),
		strings.Repeat("Y", 39),
		strings.Repeat("Z", 39),
		}
	for _, key := range invalidKeys {
		s, getErr := inst.GetSession(key)
		if getErr == nil {
			t.Errorf("Invalid key %s returned an entry with get: %v", key, s)
		}
		if _, isNoSuchSession := getErr.(gopherbouncedb.NoSuchSession); !isNoSuchSession {
			t.Errorf("Get for %s returned an error: %s", key, getErr.Error())
		}
	}

}

func TestSessionDelete(suite SessionTestSuiteBinding, t *testing.T) {
	restoreDefaultsSession()
	inst := suite.BeginInstance()
	defer suite.CloseInstance(inst)
	initErr := inst.InitSessions()
	if initErr != nil {
		t.Fatal("Init failed:", initErr)
	}
	insertSessionsOkay(inst, t)
	deletes := []*gopherbouncedb.SessionEntry{sessions[0], sessions[1]}
	for _, ds := range deletes {
		if delErr := inst.DeleteSession(ds.Key); delErr != nil {
			t.Errorf("Delete for valid key %s returned an error: %s", ds.Key, delErr.Error())
		}
	}
	// now try to look them up
	for _, ds := range deletes {
		s, getErr := inst.GetSession(ds.Key)
		_, isNotExists := getErr.(gopherbouncedb.NoSuchSession)
		if getErr == nil || !isNotExists {
			t.Errorf("Get returned an entry for key %s after the key was deleted. Entry: %v. Error: %v",
				ds.Key, s, getErr)
		}
	}
	// try to delete some invalid keys
	invalidKeys := []string{
		strings.Repeat("X", 39),
		strings.Repeat("Y", 39),
		strings.Repeat("Z", 39),
	}
	for _, key := range invalidKeys {
		if delErr := inst.DeleteSession(key); delErr != nil {
			t.Errorf("Delete for non-existing key %s returned an error: %s",
				key, delErr.Error())
		}
	}
}

func TestSessionCleanUp(suite SessionTestSuiteBinding, t *testing.T) {
	restoreDefaultsSession()
	inst := suite.BeginInstance()
	defer suite.CloseInstance(inst)
	initErr := inst.InitSessions()
	if initErr != nil {
		t.Fatal("Init failed:", initErr)
	}
	insertSessionsOkay(inst, t)
	refDate := parseTime("11-09-2019")
	numDel, cleanErr := inst.CleanUp(refDate)
	if cleanErr != nil {
		t.Errorf("CleanUp returned an error: %s", cleanErr.Error())
	}
	if numDel != 2 {
		t.Errorf("Expected CleanUp to delete 2 entries, deleted %d", numDel)
	}
	// we could test that the correct ones were delted...
}

func TestSessionDeleteForUser(suite SessionTestSuiteBinding, t *testing.T) {
	restoreDefaultsSession()
	inst := suite.BeginInstance()
	defer suite.CloseInstance(inst)
	initErr := inst.InitSessions()
	if initErr != nil {
		t.Fatal("Init failed:", initErr)
	}
	insertSessionsOkay(inst, t)
	numDel, delErr := inst.DeleteForUser(1)
	if delErr != nil {
		t.Errorf("DeleteForUser returned an error: %s", delErr.Error())
	}
	if numDel != 1 {
		t.Errorf("Expected DeleteForUser to delete one entry, deleted %d", numDel)
	}
}
