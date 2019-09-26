package main

import (
	"strings"
)

func init() {
	Register(&Getters{})
}

const ignoreTag = "@ignore"

type Getters struct {
	Scribler
}

func (b *Getters) Name() string {
	return "getter"
}

func (b *Getters) Imports(mapper Struct) map[string]string {
	return map[string]string{}
}

func (b *Getters) Generate(mapper Struct) []byte {
	for _, field := range mapper.Fields {
		if !field.HasTag(ignoreTag) {
			fieldName := field.NameOrKindName()
			b.Printf("\nfunc (t %s) %s() %s {\n", mapper.Name, strings.Title(fieldName), field.Kind.String())
			b.Printf("  return t.%s\n", fieldName)
			b.Printf("}\n")
		}
	}

	return b.Flush()
}
