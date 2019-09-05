package main

import (
	"bytes"
	"fmt"
	"strings"
)

type Builder struct {
	final bytes.Buffer
}

func (b *Builder) Name() string {
	return "builder"
}

func (b *Builder) Generate(mapper Struct) []byte {
	structName := mapper.Name + "Builder"
	b.Printf("\ntype %s struct {", structName)
	for _, field := range mapper.Fields {
		b.Printf("\n\t%s", UncapFirst(field.String()))
	}
	b.Printf("\n}\n")

	b.Printf("\nfunc New%s() %s { return %s{} }\n", structName, mapper.Name, mapper.Name)

	for _, field := range mapper.Fields {
		b.Printf("\nfunc (b *%s) %s(%s) *%s {\n", structName, strings.Title(field.Name), field.String(), structName)
		b.Printf("	b.%s = %s\n", UncapFirst(field.Name), field.Name)
		b.Printf("  return b\n")
		b.Printf("}\n")
	}

	b.Printf("\n\nfunc (b *%s) Build() %s {", structName, mapper.Name)
	b.Printf("\nreturn %s{\n", mapper.Name)
	for _, field := range mapper.Fields {
		b.Printf("%s: b.%s,\n", field.Name, UncapFirst(field.Name))
	}
	b.Printf("}\n}")

	b.Printf("\n\nfunc (src %s) ToBuild() %s {", mapper.Name, structName)
	b.Printf("\nreturn %s{\n", structName)
	for _, field := range mapper.Fields {
		b.Printf("%s: src.%s,\n", UncapFirst(field.Name), field.Name)
	}
	b.Printf("}\n}\n")

	code := b.final.Bytes()
	b.final.Reset()
	return code
}

func (g *Builder) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.final, format, args...)
}
