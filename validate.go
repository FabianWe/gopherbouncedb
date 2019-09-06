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
	"strings"
	"regexp"
	"unicode/utf8"
	"errors"
)

// PasswordValidator is a validator that checks if a given clear text password
// has a certain property (for example a certain length).
type PasswordValidator func(password string) error

// UserVerifier is a function that takes a user and returns an error if a given
// criteria isn't matched.
// For example we can check if username / email / password are given.
//
// Note that this performs validation on a user model, thus it tests the password hash
// and cannot be used to verify if a password is valid according to some other
// criteria (for example minimum length).
// A clear text password should be checked before, see PasswordValidator.
// They're also used to verify certain length properties for the database.
type UserVerifier func(u *UserModel) error

var (
	ErrEmptyUsername = errors.New("no username given")
	ErrEmptyEmail = errors.New("no EMail given")
	ErrEmptyPassword = errors.New("no password set")
	ErrInvalidEmailSyntax = errors.New("invalid syntax in email")
	ErrUsernameTooLong = errors.New("username is longer than 150 characters")
	ErrPasswordTooLong = errors.New("password is longer than 270 characters")
	ErrEmailTooLong = errors.New("email is longer than 254 characters")
	ErrFirstNameTooLong = errors.New("first name is longer than 50 characters")
	ErrLastNameTooLong = errors.New("last name is longer than 150 characters")
)

func VerifiyNameExists(u *UserModel) error {
	if strings.TrimSpace(u.Username) == "" {
		return ErrEmptyUsername
	}
	return nil
}

func VerifyEmailExists(u *UserModel) error {
	if strings.TrimSpace(u.EMail) == "" {
		return ErrEmptyEmail
	}
	return nil
}

func VerifyPasswordExists(u *UserModel) error {
	if strings.TrimSpace(u.Password) == "" {
		return ErrEmptyPassword
	}
	return nil
}

var (
	// EmailRegexp is used to verify that an email is valid.
	// It is the python version taken from https://emailregex.com/
	EmailRegexp = regexp.MustCompile(`(^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$)`)
)

func IsEmailValid(email string) error {
	if !EmailRegexp.MatchString(email) {
		return ErrInvalidEmailSyntax
	}
	return nil
}

func VerifyEmailSyntax(u *UserModel) error {
	return IsEmailValid(u.EMail)
}

func CheckUsernameLength(username string) error {
	if utf8.RuneCountInString(username)> 150 {
		return ErrUsernameTooLong
	}
	return nil
}

func CheckPasswordHashLength(password string) error {
	if utf8.RuneCountInString(password) > 270 {
		return ErrPasswordTooLong
	}
	return nil
}

func CheckEmailLength(email string) error {
	if utf8.RuneCountInString(email) > 254 {
		return ErrEmailTooLong
	}
	return nil
}

func CheckFirstNameLength(name string) error {
	if utf8.RuneCountInString(name) > 50 {
		return ErrFirstNameTooLong
	}
	return nil
}

func CheckLastNameLength(name string) error {
	if utf8.RuneCountInString(name) > 150 {
		return ErrLastNameTooLong
	}
	return nil
}

func VerifyStandardUserLengths(u *UserModel) error {
	if err := CheckUsernameLength(u.Username); err != nil {
		return err
	}
	if err := CheckPasswordHashLength(u.Password); err != nil {
		return err
	}
	if err := CheckEmailLength(u.EMail); err != nil {
		return err
	}
	if err := CheckFirstNameLength(u.FirstName); err != nil {
		return err
	}
	if err := CheckLastNameLength(u.LastName); err != nil {
		return err
	}
	return nil
}
