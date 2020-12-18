/*
Copyright 2019 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package errors

import (
	"fmt"
	"runtime"

	"github.com/pkg/errors"
)

const (
	wrapErrorMessagePrefix  string = "  --  "
	listErrorMessagePrefix  string = "  -  "
	stackTraceMessagePrefix string = "      "
)

// stack represents a stack of program counters.
type stack []uintptr

// callers returns stack of caller function
func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

// err implements error interface that has a message and stack
type err struct {
	prefix string
	msg    string
	*stack
}

// Error is implementation of error interface
func (e *err) Error() string { return e.msg }

// Format is implementation of Formater interface
func (e *err) Format(s fmt.State, verb rune) {
	message := wrapErrorMessagePrefix + e.msg
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprint(s, message)
			for i, pc := range *e.stack {
				if i > 0 {
					return
				}
				f := errors.Frame(pc)
				fmt.Fprintf(s, "\n%s%+v", e.prefix, f)
			}
			return
		}
		fallthrough
	case 's', 'q':
		fmt.Fprint(s, message)
	}
}

// wrapper implements error interface that has a message and error
type wrapper struct {
	prefix string
	msg    string
	error
}

// Error is implementation of error interface
func (w *wrapper) Error() string { return w.msg }

// Cause is implementation of causer interface
func (w *wrapper) Cause() error { return w.error }

// Format is implementation of Formater interface
func (w *wrapper) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.error)
			fmt.Fprint(s, w.prefix+w.msg)
			return
		}
		fallthrough
	case 's', 'q':
		fmt.Fprintf(s, "%s\n", w.error)
		fmt.Fprint(s, w.prefix+w.msg)
	}
}

// withStack implements error interface that has a stack and error
type withStack struct {
	prefix string
	error
	*stack
}

// Format is implementation of Formater interface
func (ws *withStack) Format(s fmt.State, verb rune) {
	message := wrapErrorMessagePrefix + fmt.Sprintf("%s", ws.error)
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprint(s, message)
			for i, pc := range *ws.stack {
				if i > 0 {
					return
				}
				f := errors.Frame(pc)
				fmt.Fprintf(s, "\n%s%+v", ws.prefix, f)
			}
			return
		}
		fallthrough
	case 's', 'q':
		fmt.Fprint(s, message)
	}
}

// Cause is implementation of causer interface
func (ws *withStack) Cause() error { return ws.error }

// ErrorList is a wrapper over list of errors
// It implements error interface
type ErrorList struct {
	Errors []error
	msg    string
}

// Error is implementation of error interface
func (el *ErrorList) Error() string {
	message := ""
	for _, err := range el.Errors {
		message += err.Error() + ":"
	}
	el.msg = message
	return message
}

// Format is implementation of Formater interface
func (el *ErrorList) Format(s fmt.State, verb rune) {
	message := ""
	for _, err := range el.Errors {
		message += "\n" + listErrorMessagePrefix + err.Error()
	}
	fmt.Fprint(s, message)

}

// WithStack annotates ErrorList with a new message and
// stack trace of caller.
func (el *ErrorList) WithStack(message string) error {
	if el == nil {
		return nil
	}
	return &withStack{
		stackTraceMessagePrefix,
		Wrap(el, message),
		callers(),
	}
}

// WithStackf annotates ErrorList with the format specifier
// and stack trace of caller.
func (el *ErrorList) WithStackf(format string, args ...interface{}) error {
	if el == nil {
		return nil
	}
	return &withStack{
		stackTraceMessagePrefix,
		Wrapf(el, format, args...),
		callers(),
	}
}
