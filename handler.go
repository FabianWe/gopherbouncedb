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
	"time"
	"strings"
)

// NoSuchUser is an error returned when the lookup of a user failed because
// no entry for that user exists.
type NoSuchUser string

// NewNoSuchUser returns a new NoSuchUser given the cause.
func NewNoSuchUser(message string) NoSuchUser {
  return NoSuchUser(message)
}

// NewNoSuchUserID returns a new NoSuchUser error with a given id.
func NewNoSuchUserID(id UserID) NoSuchUser {
  return NewNoSuchUser(fmt.Sprintf("user with id %d does not exist",id))
}

// NewNoSuchUserUsername returns a new NoSuchUser error with a given user name.
func NewNoSuchUserUsername(username string) NoSuchUser {
	return NewNoSuchUser(fmt.Sprintf("user with username \"%s\" does not exist", username))
}

// NewNoSuchUserMail returns a new NoSuchUser error with a given email.
func NewNoSuchUserMail(email string) NoSuchUser {
	return NewNoSuchUser(fmt.Sprintf("user with email \"%s\" does not exist", email))
}

// Error returns the error string.
func (e NoSuchUser) Error() string {
  return string(e)
}

// UserExists is an error returned when the creation of a user object failed
// because a user with the given credentials already exists.
type UserExists string

// NewUserExists returns a new NewUserExists given the cause.
func NewUserExists(message string) UserExists {
  return UserExists(message)
}

// Error returns the error string.
func (e UserExists) Error() string {
  return string(e)
}

// AmbiguousCredentials is an error returned when the update of an user would lead to an inconsistent database
// state, such as email already in use.
type AmbiguousCredentials string

// NewAmbiguousCredentials returns a new AmbiguousCredentials given the cause.
func NewAmbiguousCredentials(message string) AmbiguousCredentials {
	return AmbiguousCredentials(message)
}

// Error returns the error string.
func (e AmbiguousCredentials) Error() string {
	return string(e)
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

// SessionExists is the error returned if the insertion of a session failed because the
// key already exists (should rarely happen).
type SessionExists string

// NewSessionExists returns a new SessionExists error given the message.
func NewSessionExists(message string) SessionExists {
	return SessionExists(message)
}

// NewSessionExistsKey returns a new SessionExists error given the key that already
// existed in the datastore.
func NewSessionExistsKey(key string) SessionExists {
	return NewSessionExists(fmt.Sprintf("session with key %s already exists", key))
}

// Error returns the error message.
func (e SessionExists) Error() string {
	return string(e)
}

// NoSuchSession is the error returned if the lookup of a session failed because
// such a session does not exist.
type NoSuchSession string

// NewNoSuchSession returns a new NoSuchSession error given the message.
func NewNoSuchSession(message string) NoSuchSession {
	return NoSuchSession(message)
}

// NewNoSuchSessionKey returns a new NoSuchSession error given the key that doesn't exist.
func NewNoSuchSessionKey(key string) NoSuchSession {
	return NewNoSuchSession(fmt.Sprintf("session with key %s does not exist", key))
}

// Error returns the error message.
func (e NoSuchSession) Error() string {
	return string(e)
}

// UserStorage provides methods to store, retrieve, update and delete users from
// a database.
// MemdummyUserStorage provides a reference implementation but should never be used in any real code.
type UserStorage interface {
  // InitUsers should be called once to make sure all tables in the database exist etc.
  InitUsers() error
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

// SessionStorage provides methods that are used to store and deal with auth session.
//
// In general if a user gets deleted all the users' sessions should be deleted as well.
// Since we have to different interfaces there is no direct way of adapting this.
// However both storages interfaces are usually combined in a GoauthStorage,
// this way you might be able to adept to this.
// But it shouldn't be a big problem if a session for a non-existent user remains
// in the store.
type SessionStorage interface {
	// InitSessions is called once to make sure all tables and indexes exist in the database.
	InitSessions() error
	// InsertSession inserts a new session to the datastore.
	// If the session key already exists it should an error of type SessionExists.
	InsertSession(session *SessionEntry) error
	// GetSession returns the session with the given key.
	// If no such session exists it should return an error of type NoSuchSession
	GetSession(key string) (*SessionEntry, error)
	// DeleteSession deletes the session with the given key.
	// If no such session exists this will not be considered an error.
	DeleteSession(key string) error
	// CleanUp should remove all entries that are not valid any more given the
	// reference date.
	// If a session is valid should be checked with SessionEntry.IsValid.
	// It returns the number of deletes entries.
	// If the cleanup worked successfully but the driver doesn't support the number of
	// affected entries it should return an error of type NotSupported.
	CleanUp(referenceDate time.Time) (int64, error)
	// DeleteForUser deletes all session for the given user id.
	// It returns the number of deleted entries.
	// If the delete worked successfully but the driver doesn't support the number of
	// affected entries it should return an error of type NotSupported.
	DeleteForUser(user UserID) (int64, error)
}

// RetryInsertErr is returned if several inserts failed (usually with RetrySessionInsert)
// and all generated keys were invalid. This should never happen in general.
type RetryInsertErr []error

// NewRetryInsertErr returns a new RetryInsertErr given the accumulated insert errors.
func NewRetryInsertErr(errs []error) RetryInsertErr {
	return RetryInsertErr(errs)
}

// Error returns the error message.
func (e RetryInsertErr) Error() string {
	var sb strings.Builder
	sb.WriteString("failed to insert session, accumulated the following errors:")
	for _, err := range e {
		sb.WriteString("\n   ")
		sb.WriteString(err.Error())
	}
	return sb.String()
}

// RetrySessionInsert tries to insert a session key multiple times.
//
// If a key insertion failed because the key already exists we can use this method
// to create new keys and try the insert again.
// A key collision should not usually fail, thus this is function only exists
// as a precaution.
//
// This method will return all other errors (database connection failed etc.) directly
// without retrying. If the insertion failed numTries times an error of type
// RetryInsertErr is returned which contains all insertion errors.
func RetrySessionInsert(storage SessionStorage, session *SessionEntry, numTries int) error {
	var insertErr error
	errs := make([]error, 0)
	for i := 0; i < numTries; i++ {
		insertErr = storage.InsertSession(session)
		if insertErr == nil {
			return nil
		}
		// now there was some kind of failure
		if _, isSessionExists := insertErr.(SessionExists); isSessionExists {
			// append the error to our collected errors
			errs = append(errs, insertErr)
			// retry now
			// try to create a key
			newKey, keyErr := GenSessionKey()
			if keyErr != nil {
				// if key creation failed we don't actually try to insert it
				// again, we return this as a critical error
				return keyErr
			}
			// now we set the new key and the update happens in the next loop
			session.Key = newKey
		} else {
			// all other errors are returned directly
			return insertErr
		}
	}
	// we had several insert errs and couldn't insert a valid key
	return NewRetryInsertErr(errs)
}

type GoauthStorage interface {
	UserStorage
	SessionStorage
}

