package main

import (
	"fmt"
	"strings"
)

type Builder struct {
	Scribler
}

func (b *Builder) Name() string {
	return "builder"
}

func (b *Builder) Imports(mapper Struct) map[string]string {
	return map[string]string{
		"errors": "",
	}
}

func (b *Builder) Generate(mapper Struct) []byte {
	var hasRequired bool
	structName := mapper.Name + "Builder"
	b.Printf("\ntype %s struct {\n", structName)
	for _, field := range mapper.Fields {
		b.Printf("%s\n", UncapFirst(field.String()))
		if field.Required {
			b.Printf("%sDefined bool\n", UncapFirst(field.Name))
			hasRequired = true
		}
	}
	b.Printf("}\n")

	b.Printf("\nfunc New%s() %s { return %s{} }\n", structName, mapper.Name, mapper.Name)

	for _, field := range mapper.Fields {
		b.Printf("\nfunc (b *%s) %s(%s) *%s {\n", structName, strings.Title(field.Name), field.String(), structName)
		if field.Required {
			b.Printf("b.%sDefined = true\n", field.Name)
		}
		b.Printf("	b.%s = %s\n", UncapFirst(field.Name), field.Name)
		b.Printf("  return b\n")
		b.Printf("}\n")
	}

	retCode := mapper.Name
	if hasRequired {
		retCode = fmt.Sprintf("(%s, error)", retCode)
	}
	b.Printf("\n\nfunc (b *%s) Build() %s {", structName, retCode)
	for _, field := range mapper.Fields {
		if field.Required {
			b.Printf("if !b.%sDefined {\n", field.Name)
			b.Printf(" return %s{}, errors.New(\"Field %s is required.\")\n", mapper.Name, field.Name)
			b.Printf("}\n")
		}
	}
	b.Printf("\n  return %s{\n", mapper.Name)
	for _, field := range mapper.Fields {
		b.Printf("%s: b.%s,\n", field.Name, UncapFirst(field.Name))
	}
	if hasRequired {
		retCode = ", nil"
	} else {
		retCode = ""
	}
	b.Printf("  }%s\n", retCode)
	b.Printf("}\n")

	b.Printf("\n\nfunc (src %s) ToBuild() %s {", mapper.Name, structName)
	b.Printf("\nreturn %s{\n", structName)
	for _, field := range mapper.Fields {
		b.Printf("%s: src.%s,\n", UncapFirst(field.Name), field.Name)
	}
	b.Printf("}\n}\n")

	return b.Flush()
}
