# gog
code generation for Go (WIP)

generates:
* Builder - `gog:builder`
* Getters - `gog:getters`
* AllArgsContructor - `gog:allArgsContructor`
* RequiredArgsContructor - `gog:requiredArgsContructor`
* Value Objects - `gog:value`
* Record - `gog:record`

## Instalation
`go install github.com/quintans/gog@latest`

## Quick Start
Comment your struct, with the generator tag `// gog:record` and then execute `go generate ./...`.

> Using `go generate ./...` will only process files annotated with `//go:generate gog` .
> Running `gog -d <some dir>/...` will recursively scan the directories looking for go code with a recognizable tag
> `// gog:`


a source file named `src.go` with

```go
//go:generate gog

// gog:record
type Foo struct {
	// gog:@required
	name  string
	value int64
}
```

will generate a file `src_gog.go` with


```go
// Code generated by gog; DO NOT EDIT.
// Version: x.x.x
package p

import (
	"errors"
	"fmt"
)

// Generated by gog:record

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

func (f Foo) Name() string {
	return f.name
}

func (f Foo) Value() int64 {
	return f.value
}

func (f Foo) IsZero() bool {
	return f == Foo{}
}

func (f Foo) String() string {
	return fmt.Sprintf("Foo{name: %+v, value: %+v}", f.name, f.value)
}
```

(see the tests in package `plugins` for more examples)

> it is also possible to extend this create your own plugins like the one in [here](./plugins/aspect_custom_test.go)

## Guide

### gog:allArgsConstructor
generates constructor that includes all the fields

struct comment: `gog:allArgsConstructor`

field comments:
- `gog:@required` - if present validates that field is non zero

if the unexported method `validate` of the strut is present it will additionally call it as part of the constructor.
The signature is assumed to be `validate() error` 

### gog:getters
generates getters for all fields

struct comment: `gog:getters`

field comments:
- `gog:@ignore` - if present the getter for the field will not be generated

### gog:record
generates the same as `allArgsConstructor` and `getters` with the additional methods `IsZero() bool` and `String() string`. 

If any of the methods already exist in the initial struct declaration, like `IsZero() bool`, `String() string` they will not be generated.

if the unexported method `validate` of the strut is present it will additionally call it as part of the build call.
The signature is assumed to be `validate() error` 

### gog:builder
generates a builder function for the annotated struct.
No direct setter can be done on the original struct.

If an unexported setter exists it will be set the value on the target struct.
If the setter returns an error the `Build()` function will also return an error.

if the unexported method `validate` of the strut is present it will additionally call it as part of the build call.
The signature is assumed to be `validate() error` 