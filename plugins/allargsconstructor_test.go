package plugins

import (
	"fmt"
	"testing"

	"github.com/quintans/gog/config"
)

func TestAllArgsConstructor(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{
			"AllArgsConstructor",
			`
package p

// gog:allArgsConstructor
type Foo struct {
	name  string
	value int64
}
`,
			fmt.Sprintf(`// Code generated by gog; DO NOT EDIT.
// Version: %s
package p

// Generated by gog:allArgsConstructor

func NewFoo(
	name string,
	value int64,
) Foo {
	f := Foo{
		name:  name,
		value: value,
	}

	return f
}
`, config.Version),
		},
		{
			"AllArgsConstructor_with_required",
			`
package p

// gog:allArgsConstructor
type Foo struct {
	// gog:@required
	name  string
	value int64
}
`,
			fmt.Sprintf(`// Code generated by gog; DO NOT EDIT.
// Version: %s
package p

import "errors"

// Generated by gog:allArgsConstructor

func NewFoo(
	name string,
	value int64,
) (Foo, error) {
	if name == "" {
		return Foo{}, errors.New("Foo.name cannot be empty")
	}
	f := Foo{
		name:  name,
		value: value,
	}

	return f, nil
}

func MustNewFoo(
	name string,
	value int64,
) Foo {
	f, err := NewFoo(
		name,
		value,
	)
	if err != nil {
		panic(err)
	}
	return f
}
`, config.Version),
		},
		{
			"AllArgsConstructor_with_validate_method",
			`
package p

import "errors"

// gog:allArgsConstructor
type Foo struct {
	name  string
	value int64
}

func (f Foo) validate() error {
	if len(f.name) <= 3 {
		error.New("name length must be higher than 3")
	}
	return nil
}
`,
			fmt.Sprintf(`// Code generated by gog; DO NOT EDIT.
// Version: %s
package p

// Generated by gog:allArgsConstructor

func NewFoo(
	name string,
	value int64,
) (Foo, error) {
	f := Foo{
		name:  name,
		value: value,
	}
	if err := f.validate(); err != nil {
		return Foo{}, err
	}

	return f, nil
}

func MustNewFoo(
	name string,
	value int64,
) Foo {
	f, err := NewFoo(
		name,
		value,
	)
	if err != nil {
		panic(err)
	}
	return f
}
`, config.Version),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run(t, tt.in, tt.out)
		})
	}
}
