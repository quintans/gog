package vo

import (
	"errors"
	"strings"
)

type FooID struct {
	foo string
}

func NewFooID(f string) (FooID, error) {
	// silly logic
	if len(strings.Split(f, "-")) != 2 {
		return FooID{}, errors.New("unable to parse Foo ID")
	}
	return FooID{f}, nil
}
