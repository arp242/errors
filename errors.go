// Package errors adds some useful error helpers.
package errors

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	// Package to add trace lines for; if blank all traces are added.
	Package string

	// StackSize is the maximum stack sized added to errors. Set to 0 to
	// disable.
	StackSize int = 32
)

func New(text string) error                   { return addStack(errors.New(text)) }
func Errorf(f string, a ...interface{}) error { return addStack(fmt.Errorf(f, a...)) }
func Unwrap(err error) error                  { return errors.Unwrap(err) }
func Is(err, target error) bool               { return errors.Is(err, target) }
func As(err error, target interface{}) bool   { return errors.As(err, target) }

// Wrap an error with fmt.Errorf(), returning nil if err is nil.
func Wrap(err error, s string) error {
	// TODO: considerer changing this; pkg/errors people reported problems with
	// this, and actually, that makes sense.
	//
	// Instead, add errors.IfErr(err, "X") ... need to think of a better name.
	// errors.WrapIf()
	// errors.WrapIff()
	// errors.IfWrap()
	// errors.IfWrapf()
	if err == nil {
		return nil
	}
	return addStack(fmt.Errorf(s+": %w", err))
}

// Wrapf an error with fmt.Errorf(), returning nil if err is nil.
func Wrapf(err error, format string, a ...interface{}) error {
	if err == nil {
		return nil
	}
	return addStack(fmt.Errorf(format+": %w", append(a, err)...))
}

func addStack(err error) error {
	if StackSize == 0 {
		return err
	}

	pc := make([]uintptr, StackSize)
	n := runtime.Callers(3, pc)
	pc = pc[:n]

	var (
		frames = runtime.CallersFrames(pc)
		rows   = make([][]interface{}, 0, 8)
		width  = 20
	)
	for {
		f, more := frames.Next()
		if f.Function == "testing.tRunner" || f.Function == "runtime.goexit" ||
			(Package != "" && !strings.HasPrefix(f.Function, Package)) {
			if !more {
				break
			}
			continue
		}
		if !more {
			break
		}

		loc := filepath.Base(f.File) + ":" + strconv.Itoa(f.Line)
		if len(loc) > width {
			width = len(loc)
		}
		rows = append(rows, []interface{}{loc, f.Function})
	}

	// Don't format exactly the same as debug.PrintStack(); memory addresses
	// aren't very useful here and only add to the noise.
	b := new(strings.Builder)
	f := fmt.Sprintf("\t%%-%ds   %%s\n", width)
	for _, r := range rows {
		fmt.Fprintf(b, f, r...)
	}

	return &StackErr{err: err, stack: b.String()}
}

type StackTracer interface {
	StackTrace() string
}

type StackErr struct {
	stack string
	err   error
}

func (err StackErr) Unwrap() error      { return err.err }
func (err StackErr) StackTrace() string { return err.stack }

func (err StackErr) Error() string {
	if err.stack == "" {
		return fmt.Sprintf("%s", err.err)
	}
	return fmt.Sprintf("%s\n%s", err.err, err.stack)
}

// Group multiple errors.
type Group struct {
	// Maximum number of errors; calls to Append() won't do anything if the
	// number of errors is larger than this.
	MaxSize int

	mu    *sync.Mutex
	errs  []error
	nerrs int
}

// NewGroup create a new Group instance. It will record a maximum of maxSize
// errors. Set to 0 for no limit.
func NewGroup(maxSize int) *Group {
	return &Group{MaxSize: maxSize, mu: new(sync.Mutex)}
}

func (g Group) Error() string {
	if len(g.errs) == 0 {
		return ""
	}

	var b strings.Builder
	if g.nerrs > len(g.errs) {
		fmt.Fprintf(&b, "%d errors (first %d shown):\n", g.nerrs, len(g.errs))
	} else if len(g.errs) > 1 {
		fmt.Fprintf(&b, "%d errors:\n", len(g.errs))
	}
	for _, e := range g.errs {
		if e2, ok := e.(*StackErr); ok {
			e = e2.Unwrap()
		}
		b.WriteString(e.Error())
		if len(g.errs) > 1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// Len returns the number of errors we have stored.
func (g Group) Len() int { return len(g.errs) }

// Size returns the number of errors that occured.
func (g Group) Size() int { return g.nerrs }

// Append a new error to the list; this is thread-safe.
//
// It won't do anything if the error is nil, in which case it will return false.
// This makes appending errors in a loop slightly nicer:
//
//	for {
//	    err := do()
//	    if errors.Append(err) {
//	        continue
//	    }
//	}
func (g *Group) Append(err error) bool {
	if err == nil {
		return false
	}

	var gErr *Group
	errors.As(err, &gErr)

	g.mu.Lock()
	defer g.mu.Unlock()
	if gErr != nil {
		g.nerrs += len(gErr.errs)
		g.errs = append(g.errs, gErr.errs...)
	} else {
		g.nerrs++
		if g.MaxSize == 0 || len(g.errs) < g.MaxSize {
			g.errs = append(g.errs, err)
		}
	}
	return true
}

// ErrorOrNil returns itself if there are errors, or nil otherwise.
//
// It avoids an if-check at the end:
//
//	return errs.ErrorOrNil()
func (g *Group) ErrorOrNil() error {
	if g.Len() == 0 {
		return nil
	}
	return g
}

// List all the errors; returns nil if there are no errors.
func (g Group) List() []error {
	if g.Len() == 0 {
		return nil
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	e := make([]error, len(g.errs))
	copy(e, g.errs)
	return e
}
