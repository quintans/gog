package main

import (
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
	b.genStructAndNew(mapper)
	b.genBuilderSetters(mapper)
	b.genBuild(mapper)
	b.genToBuild(mapper)

	return b.Flush()
}

func (b *Builder) genStructAndNew(mapper Struct) {
	structName := mapper.Name
	b.Printf("\ntype %sBuilder struct {\n", structName)
	for _, field := range mapper.Fields {
		if field.Name == "" {
			b.Printf("%s\n", field.Kind.String())
		} else {
			b.Printf("%s\n", UncapFirst(field.String()))
		}
		if field.Required {
			b.Printf("%sDefined bool\n", UncapFirst(field.NormalizedName()))
		}
	}
	b.Printf("}\n")

	b.Printf("\nfunc New%sBuilder() *%sBuilder { return &%sBuilder{} }\n", structName, structName, structName)
}

func (b *Builder) genBuilderSetters(mapper Struct) {
	for _, field := range mapper.Fields {
		fieldName := field.NormalizedName()
		uncapFieldName := UncapFirst(fieldName)
		method := strings.Title(fieldName)
		if field.Name == "" {
			method = "With" + method
		}
		b.Printf("\nfunc (b *%sBuilder) %s(%s %s) *%sBuilder {\n", mapper.Name, method, uncapFieldName, field.Kind.String(), mapper.Name)
		if field.Required {
			b.Printf("b.%sDefined = true\n", uncapFieldName)
		}
		b.Printf("	b.%s = %s\n", fieldName, uncapFieldName)
		b.Printf("  return b\n")
		b.Printf("}\n")
	}
}

func (b *Builder) genBuild(mapper Struct) {
	structName := mapper.Name
	retCode := "*" + structName
	b.Printf("\n\nfunc (b *%sBuilder) Build() (%s, error) {\n", structName, retCode)
	for _, field := range mapper.Fields {
		if field.Required {
			fieldName := field.NormalizedName()
			uncapFieldName := UncapFirst(fieldName)
			b.Printf("if !b.%sDefined {\n", uncapFieldName)
			b.Printf(" return nil, errors.New(\"Field %s is required.\")\n", fieldName)
			b.Printf("}\n\n")
		}
	}

	b.Printf("x := &%s{}\n", structName)
	for _, field := range mapper.Fields {
		fieldName := field.NormalizedName()
		name, hasRetErr, ok := findSetterName(mapper, field)
		if ok {
			if hasRetErr {
				b.Printf("if err := ")
			}
			b.Printf("x.%s(b.%s)", name, UncapFirst(fieldName))
			if hasRetErr {
				b.Printf("; err != nil { return nil, err }\n")
			} else {
				b.Printf("\n")
			}
		} else {
			b.Printf("x.%s = b.%s\n", fieldName, UncapFirst(fieldName))
		}
	}
	b.Printf("\nreturn x, nil\n")
	b.Printf("}\n")
}

func findSetterName(mapper Struct, field Field) (string, bool, bool) {
	setter := "Set" + strings.Title(field.NormalizedName())
	for _, m := range mapper.Methods {
		if setter == strings.Title(m.Name) {
			hasRetErr := len(m.Results) == 1 && m.Results[0].Kind.Name == "error"
			return m.Name, hasRetErr, true
		}
	}
	return "", false, false
}

func (b *Builder) genToBuild(mapper Struct) {
	structName := mapper.Name
	b.Printf("\n\nfunc (src %s) ToBuild() *%sBuilder {", structName, structName)
	b.Printf("\nreturn &%sBuilder{\n", structName)
	for _, field := range mapper.Fields {
		fieldName := field.NormalizedName()
		b.Printf("%s: src.%s,\n", UncapFirst(fieldName), fieldName)
	}
	b.Printf("}\n}\n")
}
