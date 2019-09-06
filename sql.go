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
	"database/sql"
	"strings"
	"fmt"
	"time"
)

var (
	DefaultUserRowNames = map[string]string{
		"ID": "id",
		"FirstName": "first_name",
		"LastName": "last_name",
		"Username": "username",
		"EMail": "email",
		"Password": "password",
		"IsActive": "is_active",
		"IsSuperUser": "is_superuser",
		"IsStaff": "is_staff",
		"DateJoined": "date_joined",
		"LastLogin": "last_login",
	}
)

type UserSQL interface {
	Init() []string
	GetUser() string
	GetUserByName() string
	GetUserByEmail() string
	InsertUser() string
	UpdateUser(fields []string) string
	DeleteUser() string
}

type SQLTemplateReplacer struct {
	entries map[string]string
	replacer *strings.Replacer
}

func NewSQLTemplateReplacer() *SQLTemplateReplacer {
	res := &SQLTemplateReplacer{
		entries: make(map[string]string),
		replacer: nil,
	}
	res.computeReplacer()
	return res
}

func DefaultSQLReplacer() *SQLTemplateReplacer {
	res := NewSQLTemplateReplacer()
	values := map[string]string{
		"$TABLE_NAME$": "auth_user",
		"$EMAIL_UNIQUE$": "UNIQUE",
	}
	res.UpdateDict(values)
	return res
}

func (t *SQLTemplateReplacer) computeReplacer() {
	values := make([]string, 0, 2 * len(t.entries))
	for key, value := range t.entries {
		values = append(values, key, value)
	}
	t.replacer = strings.NewReplacer(values...)
}

func (t *SQLTemplateReplacer) Set(key, value string) {
	t.entries[key] = value
	t.computeReplacer()
}

func (t *SQLTemplateReplacer) SetMany(oldnew ...string) {
	if len(oldnew) % 2 != 0 {
		panic("gopherbouncedb.SQLTemplateReplacer.SetKeys: odd argument count")
	}
	for i := 0; i < len(oldnew); i += 2 {
		key, value := oldnew[i], oldnew[i+1]
		t.entries[key] = value
	}
	t.computeReplacer()
}

func (t *SQLTemplateReplacer) UpdateDict(mapping map[string]string) {
	for key, value := range mapping {
		t.entries[key] = value
	}
	t.computeReplacer()
}

func (t *SQLTemplateReplacer) Update(other *SQLTemplateReplacer) {
	t.UpdateDict(other.entries)
}

func (t *SQLTemplateReplacer) Apply(templateStr string) string {
	return t.replacer.Replace(templateStr)
}


type RollbackErr struct {
	initialErr, rollbackErr error
}

func NewRollbackErr(initialErr, rollbackErr error) RollbackErr {
	return RollbackErr{initialErr: initialErr, rollbackErr: rollbackErr}
}

func (e RollbackErr) Error() string {
	return fmt.Sprintf("Statement failed: %v, unable to rollback: %v", e.initialErr, e.rollbackErr)
}

type SQLBridge interface {
	TimeScanType() interface{}
	ConvertTimeScanType(val interface{}) (time.Time, error)
	ConvertExistsErr(err error) error
	ConvertAmbiguousErr(err error) error
	ConvertTime(t time.Time) interface{}
}

type SQLUserStorage struct {
	DB *sql.DB
	Queries UserSQL
	Bridge SQLBridge
}

func NewSQLUserStorage(db *sql.DB, queries UserSQL, bridge SQLBridge) *SQLUserStorage {
	return &SQLUserStorage{DB: db, Queries: queries, Bridge: bridge}
}

func (s *SQLUserStorage) Init() error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	// save all exec errors in a variable, return later with rollback
	var execErr error;
	for _, initQuery := range s.Queries.Init() {
		// execute only non-empty statements
		// we'll do a rollback and return that error later
		if initQuery != "" {
			if _, err := tx.Exec(initQuery); err != nil {
				execErr = err
				break
			}
		}
	}
	if execErr != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return NewRollbackErr(execErr, rollbackErr)
		}
		return execErr
	}
	// commit
	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("commit in database init failed: %s", commitErr.Error())
	}
	return nil
}

func (s *SQLUserStorage) scanUser(row *sql.Row, noUser NoSuchUser) (*UserModel, error) {
	var userId UserID
	var username, password, email, firstName, lastName string
	var isSuperuser, isStaff, isActive bool
	var dateJoined, lastLogin interface{}
	dateJoined, lastLogin = s.Bridge.TimeScanType(), s.Bridge.TimeScanType()
	scanErr := row.Scan(&userId, &username, &password, &email,
		&firstName, &lastName, &isSuperuser, &isStaff,
		&isActive, dateJoined, lastLogin)
	switch {
	case scanErr == sql.ErrNoRows:
		return nil, noUser
	case scanErr != nil:
		return nil, scanErr
	}
	var user UserModel
	user.ID = userId
	user.FirstName = firstName
	user.LastName = lastName
	user.Username = username
	user.EMail = email
	user.Password = password
	user.IsActive = isActive
	user.IsSuperUser = isSuperuser
	user.IsStaff = isStaff
	if dj, djErr := s.Bridge.ConvertTimeScanType(dateJoined); djErr != nil {
		return nil, djErr
	} else {
		user.DateJoined = dj
	}
	if ll, llErr := s.Bridge.ConvertTimeScanType(lastLogin); llErr != nil {
		return nil, llErr
	} else {
		user.LastLogin = ll
	}
	return &user, nil
}

func (s *SQLUserStorage) GetUser(id UserID) (*UserModel, error) {
	row := s.DB.QueryRow(s.Queries.GetUser(), id)
	notExists := NewNoSuchUserID(id)
	return s.scanUser(row, notExists)
}

func (s *SQLUserStorage) GetUserByName(username string) (*UserModel, error) {
	row := s.DB.QueryRow(s.Queries.GetUserByName(), username)
	notExists := NewNoSuchUserUsername(username)
	return s.scanUser(row, notExists)
}

func (s *SQLUserStorage) GetUserByEmail(email string) (*UserModel, error) {
	row := s.DB.QueryRow(s.Queries.GetUserByEmail(), email)
	notExists := NewNoSuchUserMail(email)
	return s.scanUser(row, notExists)
}

func (s *SQLUserStorage) InsertUser(user *UserModel) (UserID, error) {
	now := time.Now().UTC()
	var zeroTime time.Time
	// use the bridge conversion for time
	dateJoined := s.Bridge.ConvertTime(now)
	lastLogin := s.Bridge.ConvertTime(zeroTime)
	user.DateJoined = now
	user.LastLogin = zeroTime
	r, err := s.DB.Exec(s.Queries.InsertUser(),
		user.Username, user.Password, user.EMail, user.FirstName,
		user.LastName, user.IsSuperUser, user.IsStaff,
		user.IsActive, dateJoined, lastLogin)
	if err != nil {
		user.ID = InvalidUserID
		return InvalidUserID, s.Bridge.ConvertExistsErr(err)
	}
	lastInsertID, idErr := r.LastInsertId()
	if idErr != nil {
		// not all drivers allow this
		return InvalidUserID, NewNotSupported(idErr)
	}
	return UserID(lastInsertID), nil
}

func (s *SQLUserStorage) UpdateUser(id UserID, newCredentials *UserModel, fields []string) error {
	dateJoined := s.Bridge.ConvertTime(newCredentials.DateJoined)
	lastLogin := s.Bridge.ConvertTime(newCredentials.LastLogin)
	// this generic implementation ignores fields, thus sets fields to nil
	_, err := s.DB.Exec(s.Queries.UpdateUser(nil),
		newCredentials.Username, newCredentials.Password, newCredentials.EMail,
		newCredentials.FirstName, newCredentials.LastName, newCredentials.IsSuperUser,
		newCredentials.IsStaff, newCredentials.IsActive,
		dateJoined, lastLogin, id)
	if err != nil {
		return s.Bridge.ConvertAmbiguousErr(err)
	}
	return nil
}

func (s *SQLUserStorage) DeleteUser(id UserID) error {
	_, err := s.DB.Exec(s.Queries.DeleteUser(), id)
	return err
}
