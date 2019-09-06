package main

import (
	"strings"
)

type Getters struct {
	Scribler
}

func (b *Getters) Name() string {
	return "getter"
}

func (b *Getters) Imports(mapper Struct) map[string]string {
	return make(map[string]string)
}

func (b *Getters) Generate(mapper Struct) []byte {
	for _, field := range mapper.Fields {
		b.Printf("\nfunc (t %s) %s() %s {\n", mapper.Name, strings.Title(field.Name), field.Kind.String())
		b.Printf("  return t.%s\n", field.Name)
		b.Printf("}\n")
	}

	return b.Flush()
}