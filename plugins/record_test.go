package plugins

import (
	"fmt"
	"testing"

	"github.com/quintans/gog/config"
)

func TestRecord(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{
			"Record",
			`
package p

// gog:record
type Foo struct {
	name     string
	value    int64
	optional *string
}
`,
			fmt.Sprintf(`// Code generated by gog; DO NOT EDIT.
// Version: %s
package p

import "fmt"

// Generated by gog:record

func NewFoo(
	name string,
	value int64,
	optional *string,
) Foo {
	f := Foo{
		name:     name,
		value:    value,
		optional: optional,
	}

	return f
}

func (f Foo) Name() string {
	return f.name
}

func (f Foo) Value() int64 {
	return f.value
}

func (f Foo) Optional() *string {
	return f.optional
}

func (f Foo) IsZero() bool {
	return f.name == "" ||
		f.value == 0 ||
		f.optional == nil
}

func (f Foo) String() string {
	return fmt.Sprintf("Foo{name: %%+v, value: %%+v, optional: %%+v}", f.name, f.value, f.optional)
}
`, config.Version),
		},
		{
			"Record_with_stringer",
			`
package p

// gog:record
type Foo struct {
	// gog:@required
	name string
}

func (f Foo) String() string {
	return f.name
}
`,
			fmt.Sprintf(`// Code generated by gog; DO NOT EDIT.
// Version: %s
package p

import "errors"

// Generated by gog:record

func NewFoo(
	name string,
) (Foo, error) {
	if name == "" {
		return Foo{}, errors.New("Foo.name cannot be empty")
	}
	f := Foo{
		name: name,
	}

	return f, nil
}

func MustNewFoo(
	name string,
) Foo {
	f, err := NewFoo(
		name,
	)
	if err != nil {
		panic(err)
	}
	return f
}

func (f Foo) Name() string {
	return f.name
}

func (f Foo) IsZero() bool {
	return f == Foo{}
}
`, config.Version),
		},
		{
			"Record_with_non_primitive",
			`
package p

import "time"

// gog:record
type Foo struct {
	// gog:@required
	name  string
	clock time.Time
}
`,
			fmt.Sprintf(`// Code generated by gog; DO NOT EDIT.
// Version: %s
package p

import (
	"errors"
	"fmt"
	"time"
)

// Generated by gog:record

func NewFoo(
	name string,
	clock time.Time,
) (Foo, error) {
	if name == "" {
		return Foo{}, errors.New("Foo.name cannot be empty")
	}
	f := Foo{
		name:  name,
		clock: clock,
	}

	return f, nil
}

func MustNewFoo(
	name string,
	clock time.Time,
) Foo {
	f, err := NewFoo(
		name,
		clock,
	)
	if err != nil {
		panic(err)
	}
	return f
}

func (f Foo) Name() string {
	return f.name
}

func (f Foo) Clock() time.Time {
	return f.clock
}

func (f Foo) IsZero() bool {
	return f == Foo{}
}

func (f Foo) String() string {
	return fmt.Sprintf("Foo{name: %%+v, clock: %%+v}", f.name, f.clock)
}
`, config.Version),
		},
	}

	for _, tt := range tests {
		//! delete me
		if tt.name != "Record_with_non_primitive" {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			run(t, tt.in, tt.out)
		})
	}
}
