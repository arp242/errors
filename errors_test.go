package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
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
		/home/martin/code/errors/errors_test.go:44
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
		/home/martin/code/errors/errors_test.go:57` + "\n"
	if err.Error() != want {
		t.Errorf("\nout:  %q\nwant: %q", err.Error(), want)
	}

	Package = ""
	StackSize = 2
	err = New("err")
	want = `err
	zgo.at/errors.TestStack()
		/home/martin/code/errors/errors_test.go:67
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
