package plugins

import (
	"strings"

	"github.con/quintans/gog/generator"
)

func init() {
	generator.Register(&Builder{})
}

const RequiredTag = "@required"

type Builder struct {
	generator.Scribler
}

func (b *Builder) Name() string {
	return "builder"
}

func (b *Builder) Imports(mapper generator.Struct) map[string]string {
	return map[string]string{
		"errors": "",
	}
}

func (b *Builder) Generate(mapper generator.Struct) []byte {
	b.genStructAndNew(mapper)
	b.genBuilderSetters(mapper)
	b.genBuild(mapper)
	b.genToBuild(mapper)

	return b.Flush()
}

func (b *Builder) genStructAndNew(mapper generator.Struct) {
	structName := mapper.Name
	b.Printf("\ntype %sBuilder struct {\n", structName)
	for _, field := range mapper.Fields {
		b.Printf("%s\n", field.String())
		if field.HasTag(RequiredTag) {
			b.Printf("%sDefined bool\n", generator.UncapFirst(field.NameOrKindName()))
		}
	}
	b.Printf("}\n")

	b.Printf("\nfunc New%sBuilder() *%sBuilder { return &%sBuilder{} }\n", structName, structName, structName)
}

func (b *Builder) genBuilderSetters(mapper generator.Struct) {
	for _, field := range mapper.Fields {
		builderFieldName := builderFieldName(field)
		fieldName := field.NameOrKindName()
		method := strings.Title(fieldName)
		if field.Name == "" {
			method = "With" + method
		}
		argName := generator.UncapFirst(fieldName)
		b.Printf("\nfunc (b *%sBuilder) %s(%s %s) *%sBuilder {\n", mapper.Name, method, argName, field.Kind.String(), mapper.Name)
		if field.HasTag(RequiredTag) {
			b.Printf("b.%sDefined = true\n", builderFieldName)
		}
		b.Printf("	b.%s = %s\n", builderFieldName, argName)
		b.Printf("  return b\n")
		b.Printf("}\n")
	}
}

func (b *Builder) genBuild(mapper generator.Struct) {
	structName := mapper.Name
	retCode := "*" + structName
	b.Printf("\n\nfunc (b *%sBuilder) Build() (%s, error) {\n", structName, retCode)
	for _, field := range mapper.Fields {
		if field.HasTag(RequiredTag) {
			fieldName := field.NameOrKindName()
			uncapFieldName := generator.UncapFirst(fieldName)
			b.Printf("if !b.%sDefined {\n", uncapFieldName)
			b.Printf(" return nil, errors.New(\"Field %s is required.\")\n", fieldName)
			b.Printf("}\n\n")
		}
	}

	b.Printf("s := &%s{}\n", structName)
	for _, field := range mapper.Fields {
		fieldName := field.NameOrKindName()
		name, hasRetErr, ok := findSetterName(mapper, field)
		if ok {
			if hasRetErr {
				b.Printf("if err := ")
			}
			b.Printf("s.%s(b.%s)", name, builderFieldName(field))
			if hasRetErr {
				b.Printf("; err != nil { return nil, err }\n")
			} else {
				b.Printf("\n")
			}
		} else {
			b.Printf("s.%s = b.%s\n", fieldName, builderFieldName(field))
		}
	}
	b.Printf("\nreturn s, nil\n")
	b.Printf("}\n")
}

func builderFieldName(f generator.Field) string {
	if f.Name == "" {
		return f.Kind.String()
	}
	return generator.UncapFirst(f.Name)
}

func findSetterName(mapper generator.Struct, field generator.Field) (string, bool, bool) {
	setter := "Set" + strings.Title(field.NameOrKindName())
	for _, m := range mapper.Methods {
		if setter == strings.Title(m.Name()) {
			hasRetErr := len(m.Results) == 1 && m.Results[0].Kind.Name() == "error"
			return m.Name(), hasRetErr, true
		}
	}
	return "", false, false
}

func (b *Builder) genToBuild(mapper generator.Struct) {
	structName := mapper.Name
	b.Printf("\n\nfunc (src %s) ToBuild() *%sBuilder {", structName, structName)
	b.Printf("\nreturn &%sBuilder{\n", structName)
	for _, field := range mapper.Fields {
		fieldName := field.NameOrKindName()
		b.Printf("%s: src.%s,\n", builderFieldName(field), fieldName)
	}
	b.Printf("}\n}\n")
}