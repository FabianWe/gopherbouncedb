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

import "fmt"

// NoSuchUser is an error returned when the lookup of a user failed because
// no entry for that user exists.
type NoSuchUser struct {
  Err string
}

// NewNoSuchUser returns a new NoSuchUser given the cause.
func NewNoSuchUser(message string) NoSuchUser {
  return NoSuchUser{Err: message}
}

func NewNoSuchUserID(id UserID) NoSuchUser {
  return NewNoSuchUser(fmt.Sprintf("User with id %d does not exist",id))
}

func NewNoSuchUserUsername(username string) NoSuchUser {
	return NewNoSuchUser(fmt.Sprintf("User with username \"%s\" does not exist", username))
}

func NewNoSuchUserMail(email string) NoSuchUser {
	return NewNoSuchUser(fmt.Sprintf("User with email \"%s\" does not exist", email))
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

// AmbiguousCredentials is an error returned when the update of an user would lead to an inconsistent database
// state, such as email already in use.
type AmbiguousCredentials struct {
	Err string
}

// NewAmbiguousCredentials returns a new AmbiguousCredentials given the cause.
func NewAmbiguousCredentials(message string) AmbiguousCredentials {
	return AmbiguousCredentials{Err: message}
}

// Error returns the error string.
func (e AmbiguousCredentials) Error() string {
	return e.Err
}

// NotSupported is the error returned when inserting / updating a user and getting
// LastInsertID or RowsAffected is not supported by the driver.
type NotSupported struct {
	initial error
}

// NewNoInsertID returns a new NoInsertID.
func NewNotSupported(initial error) NotSupported {
	return NotSupported{initial: initial}
}

// Error returns the error string.
func (e NotSupported) Error() string {
	return fmt.Sprintf("LastInsertID not supported by driver: %v", e.initial)
}

// UserStorage provides methods to store, retrieve, update and delete users from
// a database.
// MemdummyUserStorage provides a reference implementation but should never be used in any real code.
type UserStorage interface {
  // Init should be called once to make sure all tables in the database exist etc.
  Init() error
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
  // If an user with the given credentials already exists (name or email, depending on which are enforced to be
  // unique) it should return InvalidUserID and an error of type UserExists.
  // The fields DateJoined is set to the current date (in UTC) and LastLogin is set to
  // the time zero value.
  // If the underlying driver does not support to get the last insert id
  // via LastInsertId InvalidUserID and an error of type NotSupported should be returned.
  // This indicates that the insertion took place but the id could not be obtained.
  InsertUser(user *UserModel) (UserID, error)
  // UpdateUser update the user with the given information, that is it uses
  // the user id to find the user and stores all new information.
  // fields is an optional argument which contains the fields to update.
  // The fields must be a subset of the UserModel attributes.
  // If given only these fields will be updated - user id is not allowed to be changed.
  // If fields is empty (nil or empty slice) all fields will be updated.
  // If the change of values would violate a consistency constraint (email or username already in use) it should not
  // update any fields but instead return an error of type AmbiguousCredentials.
  //
  // Updating a non-existing user should not lead to any error (returns nil).
  //
  // Short summary: If nil is returned everything is okay, but the user may not exist.
  // If any of the new values violates a database constraint (such as unique) AmbiguousCredentials is
  // returned.
  UpdateUser(id UserID, newCredentials *UserModel, fields []string) error
  // DeleteUser deletes the given user.
  // If no such user exists this will not be considered an error.
  DeleteUser(id UserID) error
}
