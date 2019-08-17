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
  "time"
)

// UserID is the id of a user stored in a database.
type UserID int64

// UserModel stores general information about a user, all these fields should
// be stored in the database.
//
// FirstName, LastName, Username, EMail should be self-explaining.
// Password is the hash of the password (string).
// IsActive is used as an alternative to destryoing an account.
// Because this could have some undesired effects it is preferred to just set
// the user to inactive instead of deleting the user object.
// IsSuperUser and isStaff should be true if the user is an "admin" user / part
// of the staff. This was inspired by the Django user model.
// DateJoined and LasdLogin should also be self-explaining.
// Note that LastLogin can be zero, meaning if the user never logged in
// LastLogin.IsZero() == true.
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

const (
  // InvalidUserID is used when a user id is required but no user with the
  // given credentials was found.
  InvalidUserID = UserID(-1)
)
