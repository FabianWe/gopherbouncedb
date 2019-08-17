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

// NoSuchUser is an error returned when the lookup of a user failed because
// no entry for that user exists.
type NoSuchUser struct {
  Err string
}

// NewNoSuchUser returns a new NoSuchUser given the cause.
func NewNoSuchUser(message string) NoSuchUser {
  return NoSuchUser{Err: message}
}

// Error returns the error string.
func (e NoSuchUser) Error() string {
  return e.Err
}

// UserExists is an error returned when the creation of a user object failed
// because a user with the given credentials already exists.
type UserExists struct {
  Err string
}

// NewUserExists returns a new NewUserExists given the cause.
func NewUserExists(message string) UserExists {
  return UserExists{Err: message}
}

// Error returns the error string.
func (e UserExists) Error() string {
  return e.Err
}

// UserStorage provides methods to store, retrieve, update and delete users from
// a database.
type UserStorage interface {
  // GetUser returns the user with the given id. If no such user exists it
  // should return nil and an error of type NoSuchUser.
  GetUser(id UserID) (*UserModel, error)
  // GetUserByName returns the user with the given name. If no such user exists
  // it should return nil and an error of type NoSuchUser.
  GetUserByName(username string) (*UserModel, error)
  // GetUserByEmail returns the user with the given email. If no such user
  // exists it should return nil and an error of type NoSuchUser.
  GetUserByEmail(email string) (*UserModel, error)
  // InsertUser inserts a new user to the store. It should set the id of the
  // provided user model to the new id and return that id as well.
  // If
  InsertUser(user *UserModel) (UserID, error)
  // UpdateUser update the user with the given information, that is it uses
  // the user id to find the user and stores all new information.
  // fields is an optional argument which contains the fields to update.
  // The fields must be a subset of the UserModel attributes.
  // If given only these fields will be updated.
  // If fields is empty (nil or empty slice) all fields will be updated.
  // If no such user exists it should return an error of type NoSuchUser.
  UpdateUser(user *UserModel, fields []string) error
  // DeleteUser deletes the given user.
  // If no such user exists it should return an error of type NoSuchUser.
  DeleteUser(user *UserModel) error
}
