// +build !debug

package errors

import (
	"fmt"
	"io"
	"os/exec"

	"os"
	"testing"
)

//func TestDebug(t *testing.T) {
//	// Test with -tags debug to run the tests in debug_test.go
//	cmd := exec.Command("go", "test", "-tags", "prod")
//	cmd.Stdout = os.Stdout
//	cmd.Stderr = os.Stderr
//	if err := cmd.Run(); err != nil {
//		t.Fatalf("external go test failed: %v", err)
//	}
//}

func TestMarshal(t *testing.T) {
	// Single error. No user is set, so we will have a zero-length field inside.
	e1 := E(Op("Get"), Serve, "caching in progress")

	// Nested error.
	e2 := E(Op("Read"), Undefined, e1)

	b := MarshalError(e2)
	e3 := UnmarshalError(b)

	in := e2.(*Error)
	out := e3.(*Error)

	// Compare elementwise.
	if in.Op != out.Op {
		t.Errorf("expected Op %q; got %q", in.Op, out.Op)
	}
	if in.Kind != out.Kind {
		t.Errorf("expected kind %d; got %d", in.Kind, out.Kind)
	}
	// Note that error will have lost type information, so just check its Error string.
	if in.Err.Error() != out.Err.Error() {
		t.Errorf("expected Err %q; got %q", in.Err, out.Err)
	}
}

func TestSeparator(t *testing.T) {
	defer func(prev string) {
		Separator = prev
	}(Separator)
	Separator = ":: "

	// Single error. No user is set, so we will have a zero-length field inside.
	e1 := E(Op("Get"), Serve, "Serve error")

	// Nested error.
	e2 := E(Op("Get"), Serve, e1)

	want := "Get: Serve error:: Get: Serve error"
	if errorAsString(e2) != want {
		t.Errorf("expected %q; got %q", want, e2)
	}
}

func TestDoesNotChangePreviousError(t *testing.T) {
	err := E(Serve)
	err2 := E(Op("I will NOT modify err"), err)

	expected := "I will NOT modify err: Serve error"
	if errorAsString(err2) != expected {
		t.Fatalf("Expected %q, got %q", expected, err2)
	}
	kind := err.(*Error).Kind
	if kind != Serve {
		t.Fatalf("Expected kind %v, got %v", Serve, kind)
	}
}

//func TestNoArgs(t *testing.T) {
//	defer func() {
//		err := recover()
//		if err == nil {
//			t.Fatal("E() did not panic")
//		}
//	}()
//	_ = E()
//}

type matchTest struct {
	err1, err2 error
	matched    bool
}

const (
	op  = Op("Op")
	op1 = Op("Op1")
	op2 = Op("Op2")
)

var matchTests = []matchTest{
	// Errors not of type *Error fail outright.
	{nil, nil, false},
	{io.EOF, io.EOF, false},
	{E(io.EOF), io.EOF, false},
	{io.EOF, E(io.EOF), false},
	// Success. We can drop fields from the first argument and still match.
	{E(io.EOF), E(io.EOF), true},
	{E(op, Init, io.EOF), E(op, Init, io.EOF), true},
	{E(op, Init, io.EOF, "test"), E(op, Init, io.EOF, "test", "test"), true},
	{E(op, Init), E(op, Init, io.EOF, "test", "test"), true},
	{E(op), E(op, Init, io.EOF, "test", "test"), true},
	// Failure.
	{E(io.EOF), E(io.ErrClosedPipe), false},
	{E(op1), E(op2), false},
	{E(Init), E(Serve), false},
	{E("test"), E("test1"), false},
	{E(fmt.Errorf("error")), E(fmt.Errorf("error1")), false},
	{E(op, Init, io.EOF, "test", "test1"), E(op, Init, io.EOF, "test", "test"), false},
	{E("test", Str("something")), E("test"), false}, // Test nil error on rhs.
	// Nested *Errors.
	{E(op1, E("test")), E(op1, "1", E(op2, "2", "test")), true},
	{E(op1, "test"), E(op1, "1", E(op2, "2", "test")), false},
	{E(op1, E("test")), E(op1, "1", Str(E(op2, "2", "test").Error())), false},
}

func TestMatch(t *testing.T) {
	for _, test := range matchTests {
		matched := Match(test.err1, test.err2)
		if matched != test.matched {
			t.Errorf("Match(%q, %q)=%t; want %t", test.err1, test.err2, matched, test.matched)
		}
	}
}

type kindTest struct {
	err  error
	kind Kind
	want bool
}

var kindTests = []kindTest{
	//Non-Error errors.
	{nil, Serve, false},
	{Str("not an *Error"), Serve, false},

	// Basic comparisons.
	{E(Serve), Serve, true},
	{E(Init), Serve, false},
	{E("no kind"), Serve, false},
	{E("no kind"), Logger, false},

	// Nested *Error values.
	{E("Nesting", E(Serve)), Serve, true},
	{E("Nesting", E(Logger)), Serve, false},
	{E("Nesting", E("no kind")), Serve, false},
	{E("Nesting", E("no kind")), Logger, false},
}

func TestKind(t *testing.T) {
	for _, test := range kindTests {
		got := Is(test.kind, test.err)
		if got != test.want {
			t.Errorf("Is(%q, %q)=%t; want %t", test.kind, test.err, got, test.want)
		}
	}
}

func errorAsString(err error) string {
	if e, ok := err.(*Error); ok {
		e2 := *e
		e2.stack = stack{}
		return e2.Error()
	}
	return err.Error()
}
