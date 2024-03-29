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
	"reflect"
	"testing"
	"time"
)

// TODO test ListUsers

type UserTestSuiteBinding interface {
	BeginInstance() gopherbouncedb.UserStorage
	CloseInstance(s gopherbouncedb.UserStorage)
}

var (
	users = make([]*gopherbouncedb.UserModel, 0)
)

func init() {
	restoreDefaults()
}

func restoreDefaults() {
	// setup users
	u1 := &gopherbouncedb.UserModel{}
	u2 := &gopherbouncedb.UserModel{}
	u3 := &gopherbouncedb.UserModel{}

	u1.Username = "user1"
	u1.EMail = "user1@foo.com"
	u1.FirstName = "Foo"
	u1.IsActive = true

	u2.Username = "user2"
	u2.EMail = "user2@bar.com"
	u2.IsActive = true
	u2.IsSuperUser = true
	u2.IsStaff = true

	u3.Username = "user-three"
	u3.EMail = "user3@something.org"
	u3.IsActive = true
	u3.IsSuperUser = true

	// insert on this should fail
	u4 := &gopherbouncedb.UserModel{}
	u4.Username = "user1"
	u4.EMail = "something@something.com"

	// insert should fail on this user if mail is unique
	u5 := &gopherbouncedb.UserModel{}
	u5.Username = "user5"
	u5.EMail = "user3@something.org"

	users = []*gopherbouncedb.UserModel{u1, u2, u3, u4, u5}
}

func getInsertOK() []*gopherbouncedb.UserModel {
	return users[:3]
}

func getInsertFail() []*gopherbouncedb.UserModel {
	return []*gopherbouncedb.UserModel{users[3]}
}

func getFailEmail() []*gopherbouncedb.UserModel {
	return []*gopherbouncedb.UserModel{users[4]}
}

func TestInitSuite(suite UserTestSuiteBinding, t *testing.T) {
	restoreDefaults()
	inst := suite.BeginInstance()
	defer suite.CloseInstance(inst)
	initErr := inst.InitUsers()
	if initErr != nil {
		t.Error("Init failed:", initErr)
	}
}

func insertSuccess(inst gopherbouncedb.UserStorage, t *testing.T) {
	for _, u := range getInsertOK() {
		if returnedID, insertErr := inst.InsertUser(u); insertErr != nil {
			t.Fatal("Insert failed:", insertErr.Error())
		} else {
			if u.DateJoined.IsZero() {
				t.Fatal("DateJoined not set correctly by InsertUser")
			}
			if u.ID == gopherbouncedb.InvalidUserID {
				t.Fatal("ID not set correctly by InsertUser")
			}
			if returnedID != gopherbouncedb.InvalidUserID && returnedID != u.ID {
				t.Fatalf("Storage supports insert id, but set it to wrong value: returned %d and set to %d",
					returnedID, u.ID)
			}
		}
	}
}

func TestInsertSuite(suite UserTestSuiteBinding, mailUnique bool, t *testing.T) {
	restoreDefaults()
	inst := suite.BeginInstance()
	defer suite.CloseInstance(inst)
	initErr := inst.InitUsers()
	if initErr != nil {
		t.Fatal("Init failed:", initErr)
	}
	// make inserts that should work
	insertSuccess(inst, t)
	// make inserts that should fail (with correct error)
	for _, u := range getInsertFail() {
		_, insertErr := inst.InsertUser(u)
		if insertErr == nil {
			t.Fatal("Inserting a user with duplicate username didn't fail on", u)
		}
		// error must be UserExists
		_, isUserExists := insertErr.(gopherbouncedb.UserExists)
		if !isUserExists {
			t.Fatal("Inserting a user with duplicate username didn't return UserExists, but type",
				reflect.TypeOf(insertErr), "with error message", insertErr.Error())
		}
	}
	// only test for duplicate if mailUnique is true
	if mailUnique {
		for _, u := range getFailEmail() {
			_, insertErr := inst.InsertUser(u)
			if insertErr == nil {
				t.Fatal("Inserting a user with duplicate email didn't fail on", u)
			}
			// error must be UserExists
			_, isUserExists := insertErr.(gopherbouncedb.UserExists)
			if !isUserExists {
				t.Fatal("Inserting a user with duplicate email didn't return UserExists, but type",
					reflect.TypeOf(insertErr), "with error message", insertErr.Error())
			}
		}
	}
}

func compareTime(t1, t2 time.Time) bool {
	allowedDiff := time.Duration(1 * time.Second)
	d := t1.Sub(t2)
	if d < 0 {
		d = -d
	}
	return d <= allowedDiff

}

func compareUsers(u1, u2 *gopherbouncedb.UserModel) bool {
	return u1.FirstName == u2.FirstName && u1.LastName == u2.LastName &&
		u1.Username == u2.Username && u1.EMail == u2.EMail &&
		u1.Password == u2.Password && u1.IsActive == u2.IsActive &&
		u1.IsSuperUser == u2.IsSuperUser && u1.IsStaff == u2.IsStaff &&
		compareTime(u1.DateJoined, u2.DateJoined) &&
		compareTime(u1.LastLogin, u2.LastLogin)
}

func doLookupTests(inst gopherbouncedb.UserStorage, mailUnique bool, checks []*gopherbouncedb.UserModel, t *testing.T) {
	if checks == nil {
		checks = getInsertOK()
	}
	idRes := make([]*gopherbouncedb.UserModel, 0, len(checks))
	usernameRes := make([]*gopherbouncedb.UserModel, 0, len(checks))
	emailRes := make([]*gopherbouncedb.UserModel, 0, len(checks))
	// now we should be able to lookup all three elements:
	// by id, username and email (if mail is unique)
	for _, u := range checks {
		if lookup, lookupErr := inst.GetUser(u.ID); lookupErr == nil {
			idRes = append(idRes, lookup)
		} else {
			t.Fatalf("Lookup of user with id %d returned an error: %s",
				u.ID, lookupErr.Error())
		}
		if lookup, lookupErr := inst.GetUserByName(u.Username); lookupErr == nil {
			usernameRes = append(usernameRes, lookup)
		} else {
			t.Fatalf("Lookup of user with username %s returned an error: %s",
				u.Username, lookupErr.Error())
		}
		if !mailUnique {
			continue
		}
		if lookup, lookupErr := inst.GetUserByEmail(u.EMail); lookupErr == nil {
			emailRes = append(emailRes, lookup)
		} else {
			t.Fatalf("Lookup of user with email %s returned an error: %s",
				u.EMail, lookupErr.Error())
		}
	}
	// now test for equality
	for i, u := range checks {
		lookupID := idRes[i]
		lookupName := usernameRes[i]
		var lookupMail *gopherbouncedb.UserModel
		if mailUnique {
			lookupMail = emailRes[i]
		}
		// now compare
		if !compareUsers(u, lookupID) {
			t.Errorf("ID lookup returned wrong user. Expected: %v, got %v",
				u, lookupID)
		}
		if !compareUsers(u, lookupName) {
			t.Errorf("Username lookup returned wrong user. Expected: %v, got %v",
				u, lookupName)
		}
		if mailUnique && !compareUsers(u, lookupMail) {
			t.Errorf("EMail lookup returned wrong user. Expected: %v, got %v",
				u, lookupMail)
		}
	}
}

func TestLookupSuite(suite UserTestSuiteBinding, mailUnique bool, t *testing.T) {
	restoreDefaults()
	inst := suite.BeginInstance()
	defer suite.CloseInstance(inst)
	initErr := inst.InitUsers()
	if initErr != nil {
		t.Fatal("Init failed:", initErr)
	}
	// make inserts that should work
	insertSuccess(inst, t)
	// run compares
	doLookupTests(inst, mailUnique, nil, t)

	// now perform some tests for users that shouldn't exist

	invalid1, invalid1Err := inst.GetUser(gopherbouncedb.InvalidUserID)
	invalid2, invalid2Err := inst.GetUserByEmail("user@exists.not")
	invalid3, invalid3Err := inst.GetUserByName("nonexistent")

	collected := []struct{
		u *gopherbouncedb.UserModel
		err error
		key string
	}{
		{invalid1, invalid1Err, "ID"},
		{invalid2, invalid2Err, "EMail"},
		{invalid3, invalid3Err, "Username"},
	}

	for _, res := range collected {
		u, err, key := res.u, res.err, res.key
		if err == nil {
			t.Fatalf("Lookup for %s should not succeed, but it did and returned %v",
				key, u)
		}
		// test if its NoSuchUser
		if _, isNoUserErr := err.(gopherbouncedb.NoSuchUser); !isNoUserErr {
			t.Fatalf("Get for %s for non-existing user didn't return NoSuchUser but %v with error message %s",
				key, reflect.TypeOf(err), err.Error())
		}
	}

}

func TestUpdateUserSuite(suite UserTestSuiteBinding, mailUnique bool, t *testing.T) {
	restoreDefaults()
	inst := suite.BeginInstance()
	defer suite.CloseInstance(inst)
	initErr := inst.InitUsers()
	if initErr != nil {
		t.Fatal("Init failed:", initErr)
	}
	// make inserts that should work
	insertSuccess(inst, t)
	// now perform some legal updates
	u1 := users[0]
	u1.Username = "new-user1"

	u2 := users[1]
	u2.EMail = "new@ok.de"

	u3 := users[2]
	u3.Username = "new-user3"
	u3.IsSuperUser = false

	updatesUsers := []*gopherbouncedb.UserModel{u1, u2, u3}
	// now perform updates
	// the first run performs updates with nil, the second with the fields given
	for _, u := range updatesUsers {
		if updateErr := inst.UpdateUser(u.ID, u, nil); updateErr != nil {
			t.Fatal("Update (fields not set) returned an error:", updateErr.Error())
		}
	}
	// now compare
	doLookupTests(inst, mailUnique, nil, t)

	// now updates with fields set
	updateFields := [][]string{
		{"Username"},
		{"EMail"},
		{"Username", "IsSuperUser"},
	}
	for i, u := range updatesUsers {
		if updateErr := inst.UpdateUser(u.ID, u, updateFields[i]); updateErr != nil {
			t.Fatal("Update (fields set) returned an error:", updateErr.Error())
		}
	}
	// compare again
	doLookupTests(inst, mailUnique, nil, t)
}

func TestDeleteUserSuite(suite UserTestSuiteBinding, mailUnique bool, t *testing.T) {
	restoreDefaults()
	inst := suite.BeginInstance()
	defer suite.CloseInstance(inst)
	initErr := inst.InitUsers()
	if initErr != nil {
		t.Fatal("Init failed:", initErr)
	}
	// make inserts that should work
	insertSuccess(inst, t)
	// remove user 2, then test that user 1 and 3 are still there, but 2 isn't
	u2 := users[1]
	if deleteErr := inst.DeleteUser(u2.ID); deleteErr != nil {
		t.Fatalf("Deletion of user with id %d resulted in an error: %s",
			u2.ID, deleteErr.Error())
	}
	testModels := []*gopherbouncedb.UserModel{users[0], users[2]}
	doLookupTests(inst, mailUnique, testModels, t)
	// now test that u2 isn't there any more
	u, getErr := inst.GetUser(u2.ID)
	if getErr == nil {
		t.Fatalf("Lookup for id %d should not succeed, but it did and returned %v",
			u2.ID, u)
	}
	// error should be of type NoSuchUser
	if _, isNoUserErr := getErr.(gopherbouncedb.NoSuchUser); !isNoUserErr {
		t.Fatalf("Get user for non-existing user (%d) didn't return NoSuchUser but %v with error message %s",
			u2.ID, reflect.TypeOf(getErr), getErr.Error())
	}
}
