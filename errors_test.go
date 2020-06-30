package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

var (
	_ error       = &Group{}
	_ error       = &StackErr{}
	_ StackTracer = &StackErr{}
)

func TestWrap(t *testing.T) {
	tests := []struct {
		in   error
		want string
	}{
		{Wrap(nil, "nil"), "<nil>"},
		{Wrap(errors.New("e"), ""), ": e"},
		{Wrap(errors.New("a"), "b"), "b: a"},
		{Wrap(fmt.Errorf("b: %w", errors.New("c")), "a"), "a: b: c"},

		{Wrapf(nil, "nil"), "<nil>"},
		{Wrapf(errors.New("e"), ""), ": e"},
		{Wrapf(errors.New("a"), "b"), "b: a"},
		{Wrapf(fmt.Errorf("b: %w", errors.New("c")), "a"), "a: b: c"},

		{Wrapf(errors.New("e"), "fmt: %q, %q", "X", "Y"), `fmt: "X", "Y": e`},

		{New("x"), "x"},
		{Errorf("x"), "x"},
		{Errorf("x: %w", New("y")), "x: y"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.in), func(t *testing.T) {
			out := fmt.Sprintf("%v", tt.in)
			if !strings.HasPrefix(out, tt.want) {
				t.Errorf("\nout:  %s\nwant: %s", out, tt.want)
			}
		})
	}
}

// TODO: this will only work in my env because of hard-coded path.
func TestStack(t *testing.T) {
	err := New("err")
	want := `err
	zgo.at/errors.TestStack()
		/home/martin/code/errors/errors_test.go:51
	testing.tRunner()
		/usr/lib/go/src/testing/testing.go:991
	runtime.goexit()
		/usr/lib/go/src/runtime/asm_amd64.s:1373` + "\n"
	if err.Error() != want {
		t.Errorf("\nout:  %q\nwant: %q", err.Error(), want)
	}

	Package = "zgo.at/errors"
	err = New("err")
	want = `err
	zgo.at/errors.TestStack()
		/home/martin/code/errors/errors_test.go:64` + "\n"
	if err.Error() != want {
		t.Errorf("\nout:  %q\nwant: %q", err.Error(), want)
	}

	Package = ""
	StackSize = 2
	err = New("err")
	want = `err
	zgo.at/errors.TestStack()
		/home/martin/code/errors/errors_test.go:74
	testing.tRunner()
		/usr/lib/go/src/testing/testing.go:991` + "\n"
	if err.Error() != want {
		t.Errorf("\nout:  %q\nwant: %q", err.Error(), want)
	}

	StackSize = 0
	err = New("err")
	want = `err`
	if err.Error() != want {
		t.Errorf("\nout:  %q\nwant: %q", err.Error(), want)
	}
}

func TestGroup(t *testing.T) {
	g := NewGroup(0)
	want := ""
	if g.Error() != want {
		t.Errorf("\nout:  %q\nwant: %q", g.Error(), want)
	}

	g.Append(New("X"))
	want = "1 errors:\nX\n"
	if g.Error() != want {
		t.Errorf("\nout:  %q\nwant: %q", g.Error(), want)
	}

	g.Append(fmt.Errorf("Y"))
	want = "2 errors:\nX\nY\n"
	if g.Error() != want {
		t.Errorf("\nout:  %q\nwant: %q", g.Error(), want)
	}

	// Make sure there's no race error.
	go func() { g.Append(New("A")) }()
	g.Append(New("B"))
	time.Sleep(10 * time.Millisecond)
}

func TestGroupMaxSize(t *testing.T) {
	g := NewGroup(3)
	want := ""
	if g.Error() != want {
		t.Errorf("\nout:  %q\nwant: %q", g.Error(), want)
	}

	g.Append(New("A"))
	g.Append(New("B"))
	g.Append(New("C"))
	want = "3 errors:\nA\nB\nC\n"
	if g.Error() != want {
		t.Errorf("\nout:  %q\nwant: %q", g.Error(), want)
	}

	g.Append(New("D"))
	want = "4 errors (first 3 shown):\nA\nB\nC\n"
	if g.Error() != want {
		t.Errorf("\nout:  %q\nwant: %q", g.Error(), want)
	}
}
