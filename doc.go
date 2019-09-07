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

// Package gopherbouncedb provides a database interface for gopherbounce.
// It defines an interface for accessing users and provides a generic SQL
// implementation (without importing any database drivers).
//
// Several implementations for different databases are available.
// Note that the API / interface description may change in the future if I feel
// that more functionality is required.
// So if you use your own implementation of the interfaces make sure to use a fixed
// version and upgrade if the interface changes.
//
// It uses the same approach as the SQL package: Every implementation registers with a
// unique name (using Register) and then a new handler is created with with a config
// string that is implementation depended.
package gopherbouncedb
