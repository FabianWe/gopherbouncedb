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
  "time"
	"strings"
	"fmt"
)

// UserID is the id of a user stored in a database.
type UserID int64

// UserModel stores general information about a user, all these fields should
// be stored in the database.
//
// FirstName, LastName, Username, EMail should be self-explaining.
// Password is the hash of the password (string).
// IsActive is used as an alternative to destroying an account.
// Because this could have some undesired effects it is preferred to just set
// the user to inactive instead of deleting the user object.
// IsSuperUser and IsStaff should be true if the user is an "admin" user / part
// of the staff. This was inspired by the Django user model.
// DateJoined and LastLogin should also be self-explaining.
// Note that LastLogin can be zero, meaning if the user never logged in
// LastLogin.IsZero() == true.
//
// In general UserID, Username and EMail should be unique.
//
// Because this model is usually stored in a database here is a summary of some
// conventions for the fields:
// The strings are usually varchars with the following maximum lengths:
// Username (150), password (270), EMail (254), FirstName (50), LastName(150).
// These properties can also be verified before inserting the user to a database with
// VerifyStandardUserMaxLens.
// The database implementations don't check that automatically, but the convenient
// wrappers I'm trying to implement will.
type UserModel struct {
  	ID UserID
	FirstName   string
	LastName    string
	Username    string
	EMail       string
	Password    string
	IsActive    bool
	IsSuperUser bool
	IsStaff     bool
	DateJoined  time.Time
	LastLogin   time.Time
}

func (u *UserModel) Copy() *UserModel {
	res := &UserModel{}
	res.ID = u.ID
	res.FirstName = u.FirstName
	res.LastName = u.LastName
	res.Username = u.Username
	res.EMail = u.EMail
	res.Password = u.Password
	res.IsActive = u.IsActive
	res.IsSuperUser = u.IsSuperUser
	res.IsStaff = u.IsStaff
	res.DateJoined = u.DateJoined
	res.LastLogin = u.LastLogin
	return res
}

// GetFieldByName returns the value of the field given by its string name.
//
// This helps with methods that for example only update certain fields.
// The key must be the name of one of the fields of the user model.
// If the key is invalid an error is returned.
func (u *UserModel) GetFieldByName(name string) (val interface{}, err error) {
	switch strings.ToLower(name) {
	case "id":
		val = u.ID
	case "firstname":
		val = u.FirstName
	case "lastname":
		val = u.LastName
	case "username":
		val = u.Username
	case "email":
		val = u.EMail
	case "password":
		val = u.Password
	case "isactive":
		val = u.IsActive
	case "issuperuser":
		val = u.IsSuperUser
	case "isstaff":
		val = u.IsStaff
	case "datejoined":
		val = u.DateJoined
	case "lastlogin":
		val = u.LastLogin
	default:
		err = fmt.Errorf("invalid field name \"%s\": Must be a valid field name of the user model", name)
	}
	return
}

const (
  // InvalidUserID is used when a user id is required but no user with the
  // given credentials was found.
  InvalidUserID = UserID(-1)
)
