package main

import (
	"bytes"
	"fmt"
	"strings"
)

type Getters struct {
	final bytes.Buffer
}

func (b *Getters) Name() string {
	return "getter"
}

func (b *Getters) Generate(mapper Struct) []byte {
	for _, field := range mapper.Fields {
		b.Printf("\nfunc (t %s) %s() %s {\n", mapper.Name, strings.Title(field.Name), field.Kind.String())
		b.Printf("  return t.%s\n", field.Name)
		b.Printf("}\n")
	}

	code := b.final.Bytes()
	b.final.Reset()
	return code
}

func (g *Getters) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.final, format, args...)
}
