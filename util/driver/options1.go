// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package driver provides an external entry point to the pprof driver.
package driver

import (
	"io"
	"time"

	"github.com/sudachen/pprof/internal/plugin"
	"github.com/sudachen/pprof/profile"
)

type HTTPServerArgs = plugin.HTTPServerArgs

// Options groups all the optional plugins into pprof.
type Options struct {
	Writer  Writer
	Flagset FlagSet
	Fetch   Fetcher
	UI      UI

	HTTPServer  func(args *HTTPServerArgs) error
	WantBrowser bool

	internal plugin.Options
}

// Writer provides a mechanism to write data under a certain name,
// typically a filename.
type Writer interface {
	Open(name string) (io.WriteCloser, error)
}

// A FlagSet creates and parses command-line flags.
// It is similar to the standard flag.FlagSet.
type FlagSet interface {
	// Bool, Int, Float64, and String define new flags,
	// like the functions of the same name in package flag.
	Bool(name string, def bool, usage string) *bool
	Int(name string, def int, usage string) *int
	Float64(name string, def float64, usage string) *float64
	String(name string, def string, usage string) *string

	// BoolVar, IntVar, Float64Var, and StringVar define new flags referencing
	// a given pointer, like the functions of the same name in package flag.
	BoolVar(pointer *bool, name string, def bool, usage string)
	IntVar(pointer *int, name string, def int, usage string)
	Float64Var(pointer *float64, name string, def float64, usage string)
	StringVar(pointer *string, name string, def string, usage string)

	// StringList is similar to String but allows multiple values for a
	// single flag
	StringList(name string, def string, usage string) *[]*string

	// ExtraUsage returns any additional text that should be
	// printed after the standard usage message.
	// The typical use of ExtraUsage is to show any custom flags
	// defined by the specific pprof plugins being used.
	ExtraUsage() string

	// Parse initializes the flags with their values for this run
	// and returns the non-flag command line arguments.
	// If an unknown flag is encountered or there are no arguments,
	// Parse should call usage and return nil.
	Parse(usage func()) []string
}

// A Fetcher reads and returns the profile named by src, using
// the specified duration and timeout. It returns the fetched
// profile and a string indicating a URL from where the profile
// was fetched, which may be different than src.
type Fetcher interface {
	Fetch(src string, duration, timeout time.Duration) (*profile.Profile, string, error)
}

// A UI manages user interactions.
type UI interface {
	// Read returns a line of text (a command) read from the user.
	// prompt is printed before reading the command.
	ReadLine(prompt string) (string, error)

	// Print shows a message to the user.
	// It formats the text as fmt.Print would and adds a final \n if not already present.
	// For line-based UI, Print writes to standard error.
	// (Standard output is reserved for report data.)
	Print(...interface{})

	// PrintErr shows an error message to the user.
	// It formats the text as fmt.Print would and adds a final \n if not already present.
	// For line-based UI, PrintErr writes to standard error.
	PrintErr(...interface{})

	// IsTerminal returns whether the UI is known to be tied to an
	// interactive terminal (as opposed to being redirected to a file).
	IsTerminal() bool

	// SetAutoComplete instructs the UI to call complete(cmd) to obtain
	// the auto-completion of cmd, if the UI supports auto-completion at all.
	SetAutoComplete(complete func(string) string)
}

type WebUI interface {
	WebServer(*plugin.HTTPServerArgs) error
}
