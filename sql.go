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
	"reflect"
)

var (
	// DefaultUserRowNames matches the fields from UserModel (as strings)
	// to the default name of a sql row.
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

// UserSQL defines an interface for working with user queries in a sql database.
//
// It is used in the generic SQLStorage to retrieve database specific queries.
// The queries may (and should) contain placeholders (not the queries returned by
// your implementation, see the note below).
// For example the table name might be changed, the default name for the user table
// is "auth_user". To be more flexible this, the table name can be changed.
// Thus the queries can contain a variable that gets replaced with the actual
// table name. This meta variable has the form $SOME_NAME$.
// The following variables are enabled by default:
// "$USERS_TABLE_NAME$": Name of the users table. Defaults to "auth_user".
// "$EMAIL_UNIQUE$": Specifies if the E-Mail should be unique.
// By default it is set to the string "UNIQUE". But it can be replaced by an empty string
// as well. This should be fine with most sql implementations.
// If not you might write your own implementation that does something different and does
// not use "$EMAIL_UNIQUE$".
//
// The replacement of the meta variables should only done once during the initialization.
// A SQLTemplateReplacer is used to achieve this.
//
// Important note: The queries returned by this implementation are not allowed to contain
// meta variables. A replacer is not run by default!
// Instead you have to create the queries with placeholders once (for example as constants)
// and then apply a replacer by yourself once to get rid of the placeholders.
// Of course you don't need to use this feature, but it keeps your tables more dynamic
// and allows more configuration.
// As an example you might look at one of the implementations, for example the driver.
// All queries exist as a const string with placeholders.
// Then a replacer is run once and the implementation only returns those strings.
// They also use other placeholders to be used for example with for dynamic update queries.
// The replacement of the fields variables is then done directly in the UpdateUser query.
type UserSQL interface {
	// InitUsers returns a sequence of init actions.
	// They're all run on one transaction and rolled-back if one fails.
	// In this query als initialization should happen, usually something like
	// "create table if not exists" (or creating an index).
	InitUsers() []string
	// GetUser is the query to return a user with a given id.
	// It must select all fields from the user table in the following order:
	// id, user name, password, email, first name, last name, is superuser,
	// is staff, is active, date joined, last login.
	//
	// Exactly one element is passed to the query and that is the user id to look for.
	GetUser() string
	// GetUserByName does the same as GetUser but instead of an id gets a user name to
	// look for.
	GetUserByName() string
	// GetUserByEmail does the same as GetUser but instead of an id gets an email to
	// look for.
	// If the email is not unique this might lead to errors.
	GetUserByEmail() string
	// InsertUser inserts a new user into the database.
	//
	// The arguments parsed into Execute are the same once (and in the same order)
	// as in GetUser, except the id field (that is automatically generated).
	InsertUser() string
	// UpdateUser is used to update a user.
	// The query might depend on the fields which we want to update.
	// You don't have to support update by fields and just update all fields, even
	// those not given in fields. Just make sure to implement this in the
	// SupportsUserFields function.
	//
	// The sql implementation works as follows: If SupportsUserFields returns false
	// UpdateUser is always called with nil, meaning all fields must be updated.
	// If SupportsUserFields returns true then this function is called with the
	// actual fields and the returned statement updates should only those fields.
	//
	// Concerning in the arguments: In case len(fields) == 0 the same order as in GetUser,
	// except the id (this one can't be updated).
	// If fields is given the order of the arguments are in the same order as the fields.
	// The contents of fields are discussed in more detail in the documentation of the
	// UserStorage interface.
	// In all cases the id is passed as the last element.
	// This is the argument used usually in the WHERE clause and defines the user to
	// update by its id.
	//
	// A small example: If fields is nil: ... SET username=?, ... WHERE id=?.
	// The arguments are given in the order username, ..., id.
	//
	// If fields is given for example as ["LastName", "EMail"]:
	// ... SET last_name=?, email=? WHERE id=?.
	// The arguments are given in the order last_name, email, id.
	UpdateUser(fields []string) string
	// SupportsUserFields returns true if UpdateUser has the additional fields update
	// ability.
	// It's totally okay to return false and always update all values.
	// In this case UpdateUser always gets called with nil.
	SupportsUserFields() bool
	// DeleteUser removes a user from the database.
	// It is given a single argument, the user id.
	DeleteUser() string
}

// SQLTemplateReplacer is used to replace the meta variables in the queries of a UserSQL
// implementation.
// It basically maps these meta variable to their actual content and offers a method to
// apply the replacement.
//
// The Apply method is safe to be called concurrently, all functions that in some way
// change the content are not safe to be called concurrently.
//
// That is: First set the content and then use Apply as you see fit.
type SQLTemplateReplacer struct {
	entries map[string]string
	replacer *strings.Replacer
}

// NewSQLTemplateReplacer returns a new SQLTemplateReplacer with no replacements taking
// place.
// DefaultSQLReplacer should be used to generate a replacer with the default replacements
// taking place.
func NewSQLTemplateReplacer() *SQLTemplateReplacer {
	res := &SQLTemplateReplacer{
		entries: make(map[string]string),
		replacer: nil,
	}
	res.computeReplacer()
	return res
}

// DefaultSQLReplacer returns a new SQLTemplateReplacer that takes care that all variables
// mentioned in the documentation of UserSQL are mapped to their default values.
func DefaultSQLReplacer() *SQLTemplateReplacer {
	res := NewSQLTemplateReplacer()
	values := map[string]string{
		"$USERS_TABLE_NAME$": "auth_user",
		"$EMAIL_UNIQUE$": "UNIQUE",
	}
	res.UpdateDict(values)
	return res
}

// computeReplacer computes the new strings.Replacer.
// This method is called when the content is changed in some way.
func (t *SQLTemplateReplacer) computeReplacer() {
	values := make([]string, 0, 2 * len(t.entries))
	for key, value := range t.entries {
		values = append(values, key, value)
	}
	t.replacer = strings.NewReplacer(values...)
}

// HasKey returns true if there exists an entry for the given key.
func (t *SQLTemplateReplacer) HasKey(key string) bool {
	_, has := t.entries[key]
	return has
}

// Set sets the meta variable to a new value.
func (t *SQLTemplateReplacer) Set(key, value string) {
	t.entries[key] = value
	t.computeReplacer()
}

// SetMany sets many key / value pairs.
// It should be a bit more efficient than calling Set for each entry.
// oldnew must be a sequence with entries of the form [KEY_ONE, VALUE_ONE, KEY_TWO, VALUE_TWO, ...].
//
// All entries not mentioned in oldnew are not changed and not deleted.
//
// It panics if given an odd number of arguments
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

// UpdateDict is another way to update the key / value mapping.
// All entries contained in mapping are updated, all other entries are not changed and
// not deleted.
func (t *SQLTemplateReplacer) UpdateDict(mapping map[string]string) {
	for key, value := range mapping {
		t.entries[key] = value
	}
	t.computeReplacer()
}

// Update updates the entries by updating the fields from another replacer.
// It works the same way as UpdateDict.
func (t *SQLTemplateReplacer) Update(other *SQLTemplateReplacer) {
	t.UpdateDict(other.entries)
}

// Delete removes an entry from the mapping.
// If the key is not present nothing happens.
func (t *SQLTemplateReplacer) Delete(key string) {
	delete(t.entries, key)
	t.computeReplacer()
}

// DeleteMany deletes multiple keys from the mapping, it is more efficient than to call
// Delete for each entry.
// If a key is not present nothing happens.
func (t *SQLTemplateReplacer) DeleteMany(keys ...string) {
	for _, key := range keys {
		delete(t.entries, key)
	}
	t.computeReplacer()
}

// Apply replaces all meta variables that are a key in the template string with
// their respective values.
func (t *SQLTemplateReplacer) Apply(templateStr string) string {
	return t.replacer.Replace(templateStr)
}

// RollbackErr is returned when the rollback operation of a transaction failed.
// This is usually not a good sign: There was an error before and we tried to
// rollback the operations already performed. But this rollback resulted in yet
// another error.
type RollbackErr struct {
	initialErr, rollbackErr error
}

// NewRollbackErr returns a new RollbackErr given the cause that lead to calling
// rollback (initialErr) and the rollback error (rollbackErr) itself.
func NewRollbackErr(initialErr, rollbackErr error) RollbackErr {
	return RollbackErr{initialErr: initialErr, rollbackErr: rollbackErr}
}

// Error returns the error string.
func (e RollbackErr) Error() string {
	return fmt.Sprintf("statement failed: %s, unable to rollback: %s",
		e.initialErr.Error(), e.rollbackErr.Error())
}

// SQLBridge is a type that is used to abstract away certain driver specific implementation
// problems.
//
// This might not be the "best" approach, but it is one that works.
// Some of the things in database/sql are not very generic.
// For example various drivers handle time.Time differently.
// Also for certain errors (such as duplicate key errors) there are no generic
// error types. To have more detailed control this bridge is used to deal with these problems.
type SQLBridge interface {
	// TimeScanType should return the type that is used to retrieve
	// time.Time objects from the database.
	// When we retrieve for example a user a variable is created with this function
	// and then passed to the scan method to retrieve a time from the database.
	// Thus it should return a pointer s.t. the database Scan method can
	// assign it the actual value.
	//
	// After the retrieving with Scan is done this object is converted to a time.Time
	// with ConvertTimeScanType.
	//
	// The easiest implementation is to just return a *time.Time.
	TimeScanType() interface{}
	// ConvertTimeScanType is used to transform the values that were processed with
	// a variable from TimeScanType, thus this function can assume that val is of
	// the type returned by TimeScanType.
	// However, type checking should be done and an error returned if this is not the case.
	// Thus the workflow for retrieving time.Time elements from the database is as followś:
	// Call database Scan method with the value retrieved from TimeScanType.
	// This value is then converted to an actual time.Time with this function.
	//
	// For example if TimeScanType returns x = *time.Time this method can just return
	// *x.
	ConvertTimeScanType(val interface{}) (time.Time, error)
	// IsDuplicateInsert checks if the error is an error that was caused by inserting
	// a duplicate entry.
	//
	// Various database drivers have their own way of defining such a key error, for
	// example by an error code or a specific error type.
	IsDuplicateInsert(err error) bool
	// IsDuplicateUpdate is used the same way as IsDuplicateInsert, but is used in
	// update operations.
	// Usually database drivers don't distinguish between different key errors
	// on insert/update and thus in most cases it works the same way as IsDuplicateInsert.
	IsDuplicateUpdate(err error) bool
	// ConvertTime has the same idea as TimeScanType: Transform an entry of time.Time
	// to a driver specific time that can be used for this driver.
	// Whereas TimeScanType is used for Scan the value returned by ConvertTime
	// is used on inserts and updates.
	// For example a driver may not be able to insert a time.Time value into the database
	// directly. Instead it may have to be converted to a string instead.
	//
	// In contrast to TimeScanType it should however (in general) not return a pointer.
	// A driver that can insert time.Time directly should simply returned the supplied
	// argument of time.Time.
	ConvertTime(t time.Time) interface{}
}

// SQLUserStorage implements UserStorage by working with database/sql.
//
// It does not rely on a specific driver and no driver is imported; it only uses
// methods like db.Scan or db.Execute.
//
// In order to use your own implementation for these generic sql methods two things
// must be implemented: The queries to be used of type UserSQL and the database bridge
// of type SQLBridge.
type SQLUserStorage struct {
	DB *sql.DB
	Queries UserSQL
	Bridge SQLBridge
}

// NewSQLUserStorage returns a new SQLUserStorage.
func NewSQLUserStorage(db *sql.DB, queries UserSQL, bridge SQLBridge) *SQLUserStorage {
	return &SQLUserStorage{DB: db, Queries: queries, Bridge: bridge}
}

// InitUsers executes all init queries in a single transaction.
func (s *SQLUserStorage) InitUsers() error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	// save all exec errors in a variable, return later with rollback
	var execErr error
	for _, initQuery := range s.Queries.InitUsers() {
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

func (s *SQLUserStorage) scanUser(row *sql.Row, noUser func() error) (*UserModel, error) {
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
		return nil, noUser()
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
		dj = dj.UTC()
		user.DateJoined = dj
	}
	if ll, llErr := s.Bridge.ConvertTimeScanType(lastLogin); llErr != nil {
		return nil, llErr
	} else {
		ll = ll.UTC()
		user.LastLogin = ll
	}
	return &user, nil
}

func (s *SQLUserStorage) GetUser(id UserID) (*UserModel, error) {
	row := s.DB.QueryRow(s.Queries.GetUser(), id)
	notExists := func() error {
		return NewNoSuchUserID(id)
	}
	return s.scanUser(row, notExists)
}

func (s *SQLUserStorage) GetUserByName(username string) (*UserModel, error) {
	row := s.DB.QueryRow(s.Queries.GetUserByName(), username)
	notExists := func() error {
		return NewNoSuchUserUsername(username)
	}
	return s.scanUser(row, notExists)
}

func (s *SQLUserStorage) GetUserByEmail(email string) (*UserModel, error) {
	row := s.DB.QueryRow(s.Queries.GetUserByEmail(), email)
	notExists := func() error {
		return NewNoSuchUserMail(email)
	}
	return s.scanUser(row, notExists)
}

func (s *SQLUserStorage) InsertUser(user *UserModel) (UserID, error) {
	user.ID = InvalidUserID
	now := time.Now().UTC()
	var zeroTime time.Time
	zeroTime = zeroTime.UTC()
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
		if s.Bridge.IsDuplicateInsert(err) {
			return InvalidUserID,
				NewUserExists(fmt.Sprintf("unique constraint failed: %s", err.Error()))
		}
		return InvalidUserID, err
	}
	lastInsertID, idErr := r.LastInsertId()
	if idErr != nil {
		// not all drivers allow this
		return InvalidUserID, NewNotSupported(idErr)
	}
	user.ID = UserID(lastInsertID)
	return UserID(lastInsertID), nil
}

func (s *SQLUserStorage) prepareUpdateArgs(id UserID, u *UserModel, fields []string) ([]interface{}, error) {
	var res []interface{}
	if len(fields) == 0 {
		dateJoined := s.Bridge.ConvertTime(u.DateJoined.UTC())
		lastLogin := s.Bridge.ConvertTime(u.LastLogin.UTC())
		res = []interface{}{
			u.Username, u.Password, u.EMail, u.FirstName, u.LastName, u.IsSuperUser,
			u.IsStaff, u.IsActive, dateJoined, lastLogin,
			id,
		}
	} else {
		res = make([]interface{}, len(fields)+1)
		for i, fieldName := range fields {
			if arg, argErr := u.GetFieldByName(fieldName); argErr == nil {
				fieldName = strings.ToLower(fieldName)
				if fieldName == "datejoined" || fieldName == "lastlogin" {
					if t, isTime := arg.(time.Time); isTime {
						arg = s.Bridge.ConvertTime(t.UTC())
					} else {
						return nil,
							fmt.Errorf("DateJoined / LastLogin must be time.Time, got type %v", reflect.TypeOf(arg))
					}
				}
				res[i] = arg
			} else {
				return nil, argErr
			}
		}
		res[len(fields)] = id
	}
	return res, nil
}

// UpdateUser is a bit more complicated then the other functions.
// Its behavior is defined in the UserSQL documentation, but here's a small
// summary of what happens: If SupportsUserFields returns false the query call from
// UpdateUser(nil) is used and the arguments are given in the default order, that is
// username, email, ... and the id as the last element.
//
// If SupportsUserFields returns true only the arguments as defined in fields are
// given (in the order as tehy're mentioned) and UpdateUser is called with these fields
// and must return a query that updates these fields.
// Again the user id is given as the last argument.
func (s *SQLUserStorage) UpdateUser(id UserID, newCredentials *UserModel, fields []string) error {
	// check if it's supported to use fields, compute actual arguments depending on that
	var stmt string
	var args []interface{}
	var argsErr error
	if s.Queries.SupportsUserFields() {
		stmt = s.Queries.UpdateUser(fields)
		args, argsErr = s.prepareUpdateArgs(id, newCredentials, fields)
	} else {
		stmt = s.Queries.UpdateUser(nil)
		args, argsErr = s.prepareUpdateArgs(id, newCredentials, nil)
	}
	if argsErr != nil {
		return fmt.Errorf("can't prepare user update arguments: %s", argsErr.Error())
	}

	_, err := s.DB.Exec(stmt, args...)
	if err != nil {
		if s.Bridge.IsDuplicateUpdate(err) {
			return NewAmbiguousCredentials(fmt.Sprintf("unique constraint failed: %s", err.Error()))
		}
		return err
	}
	return nil
}

func (s *SQLUserStorage) DeleteUser(id UserID) error {
	_, err := s.DB.Exec(s.Queries.DeleteUser(), id)
	return err
}
