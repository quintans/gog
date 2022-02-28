package plugins

import (
	"strings"

	"github.com/quintans/gog/generator"
)

func init() {
	generator.Register(&Builder{})
}

type BuilderOptions struct{}

type Builder struct {
	generator.Scribler
}

func (b *Builder) Name() string {
	return "builder"
}

func (b *Builder) Imports(mapper generator.Struct) map[string]string {
	return map[string]string{}
}

func (b *Builder) GenerateBody(mapper generator.Struct) error {
	return b.WriteBody(mapper, BuilderOptions{})
}

func (b *Builder) WriteBody(mapper generator.Struct, _ BuilderOptions) error {
	b.genStructAndNew(mapper)
	b.genBuilderSetters(mapper)
	b.genBuild(mapper)
	b.genToBuild(mapper)
	b.genGetters(mapper)

	_ = PrintIsZero(&b.Scribler, mapper)
	_ = PrintString(&b.Scribler, mapper)

	return nil
}

func (b *Builder) genStructAndNew(mapper generator.Struct) {
	structName := mapper.Name
	b.BPrintf("\ntype %sBuilder struct {\n", structName)
	for _, field := range mapper.Fields {
		b.BPrintf("%s\n", field.String())
	}
	b.BPrintf("}\n")

	args := &generator.Scribler{}
	props := &generator.Scribler{}
	for _, field := range mapper.Fields {
		if field.HasTag(RequiredTag) {
			name := generator.UncapFirst(field.NameOrKindName())
			args.BPrintf("%s %s,", name, field.Kind.String())
			props.BPrintf("%s: %s,\n", name, name)
		}
	}
	b.BPrintf("\nfunc New%sBuilder(%s) *%sBuilder {\n return &%sBuilder{\n%s} \n}\n", structName, args, structName, structName, props)
}

func (b *Builder) genBuilderSetters(mapper generator.Struct) {
	for _, field := range mapper.Fields {
		builderFieldName := field.NameForField()
		fieldName := field.NameOrKindName()
		method := strings.Title(fieldName)
		if field.Name == "" {
			method = "With" + method
		}
		argName := generator.UncapFirst(fieldName)
		b.BPrintf("\nfunc (b *%sBuilder) %s(%s %s) *%sBuilder {\n", mapper.Name, method, argName, field.Kind.String(), mapper.Name)
		b.BPrintf("	b.%s = %s\n", builderFieldName, argName)
		b.BPrintf("  return b\n")
		b.BPrintf("}\n")
	}
}

func (b *Builder) genBuild(mapper generator.Struct) {
	s := &generator.Scribler{}
	hasError := PrintZeroCheck(s, mapper, "b")

	structName := mapper.Name
	s.BPrintf("s := %s{\n", structName)
	for _, field := range mapper.Fields {
		fieldName := field.NameOrKindName()
		s.BPrintf("	%s: b.%s,\n", fieldName, field.NameForField())
	}
	s.BPrintf("  }\n\n")

	hasError = PrintValidate(s, mapper, "s") || hasError
	retCode := structName
	if hasError {
		retCode = "(" + retCode + ", error)"
	}
	b.BPrintf("\n\nfunc (b *%sBuilder) Build() %s {\n", structName, retCode)
	b.BPrintf("%s\n", s)
	b.BPrintf("return s")
	if hasError {
		b.BPrintf(", nil")
	}
	b.BPrintf("\n}\n")
}

func (b *Builder) genToBuild(mapper generator.Struct) {
	structName := mapper.Name
	b.BPrintf("\n\nfunc (b *%s) ToBuild() *%sBuilder {", structName, structName)
	b.BPrintf("\nreturn &%sBuilder{\n", structName)
	for _, field := range mapper.Fields {
		fieldName := field.NameOrKindName()
		b.BPrintf("%s: b.%s,\n", field.NameForField(), fieldName)
	}
	b.BPrintf("}\n}\n")
}

func (b *Builder) genGetters(mapper generator.Struct) {
	getters := Getters{}
	getters.WriteBody(mapper, GetterOptions{})
	b.BPrintf("\n")
	b.Body.Write(getters.Body.Bytes())
}
