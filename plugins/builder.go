package plugins

import (
	"strings"

	"github.com/quintans/gog/generator"
)

func init() {
	generator.Register(&Builder{})
}

const RequiredTag = "@required"

type BuilderOptions struct {
	Star bool
}

type Builder struct {
	generator.Scribler
}

func (b *Builder) Name() string {
	return "builder"
}

func (b *Builder) Imports(mapper generator.Struct) map[string]string {
	return map[string]string{
		"\"errors\"": "",
	}
}

func (b *Builder) Generate(mapper generator.Struct) ([]byte, error) {
	options := BuilderOptions{
		Star: true,
	}
	if tag, ok := mapper.FindTag(b.Name()); ok {
		if err := tag.Unmarshal(&options); err != nil {
			return nil, err
		}
	}

	b.genStructAndNew(mapper)
	b.genBuilderSetters(mapper)
	b.genBuild(mapper, options)
	b.genToBuild(mapper)

	return b.Flush(), nil
}

func (b *Builder) genStructAndNew(mapper generator.Struct) {
	structName := mapper.Name
	b.Printf("\ntype %sBuilder struct {\n", structName)
	for _, field := range mapper.Fields {
		b.Printf("%s\n", field.String())
	}
	b.Printf("}\n")

	args := &generator.Scribler{}
	props := &generator.Scribler{}
	for _, field := range mapper.Fields {
		if field.HasTag(RequiredTag) {
			name := generator.UncapFirst(field.NameOrKindName())
			args.Printf("%s %s,", name, field.Kind.String())
			props.Printf("%s: %s,\n", name, name)
		}
	}
	b.Printf("\nfunc New%sBuilder(%s) *%sBuilder {\n return &%sBuilder{\n%s} \n}\n", structName, args, structName, structName, props)
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
		b.Printf("	b.%s = %s\n", builderFieldName, argName)
		b.Printf("  return b\n")
		b.Printf("}\n")
	}
}

func (b *Builder) genBuild(mapper generator.Struct, options BuilderOptions) {
	structName := mapper.Name

	s := &generator.Scribler{}
	var retCode string
	if options.Star {
		retCode = "*"
		s.Printf("s := &%s{}\n", structName)
	} else {
		s.Printf("s := %s{}\n", structName)
	}
	retCode += structName

	var hasError bool
	for _, field := range mapper.Fields {
		fieldName := field.NameOrKindName()
		name, hasRetErr, ok := findSetterName(mapper, field)
		if ok {
			if hasRetErr {
				s.Printf("if err := ")
			}
			s.Printf("s.%s(b.%s)", name, builderFieldName(field))
			if hasRetErr {
				s.Printf("; err != nil { return nil, err }\n")
				hasError = true
			} else {
				s.Printf("\n")
			}
		} else {
			s.Printf("s.%s = b.%s\n", fieldName, builderFieldName(field))
		}
	}
	if hasError {
		retCode = "(" + retCode + ", error)"
	}
	b.Printf("\n\nfunc (b *%sBuilder) Build() %s {\n", structName, retCode)
	b.Printf("%s\n", s)
	b.Printf("return s")
	if hasError {
		b.Printf(", nil")
	}
	b.Printf("\n}\n")
}

func builderFieldName(f generator.Field) string {
	if f.Name == "" {
		return f.Kind.String()
	}
	return generator.UncapFirst(f.Name)
}

func findSetterName(mapper generator.Struct, field generator.Field) (string, bool, bool) {
	setter := "Set" + strings.Title(field.NameOrKindName())
	if m, ok := mapper.FindMethod(setter); ok {
		hasRetErr := len(m.Results) == 1 && m.Results[0].Kind.Name() == "error"
		return m.Name(), hasRetErr, true
	}
	return "", false, false
}

func (b *Builder) genToBuild(mapper generator.Struct) {
	structName := mapper.Name
	b.Printf("\n\nfunc (src *%s) ToBuild() *%sBuilder {", structName, structName)
	b.Printf("\nreturn &%sBuilder{\n", structName)
	for _, field := range mapper.Fields {
		fieldName := field.NameOrKindName()
		b.Printf("%s: src.%s,\n", builderFieldName(field), fieldName)
	}
	b.Printf("}\n}\n")
}
