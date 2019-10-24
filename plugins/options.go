package plugins

import (
	"strings"

	"github.com/quintans/gog/generator"
)

func init() {
	generator.Register(&Options{})
}

type Options struct {
	generator.Scribler
}

func (b *Options) Name() string {
	return "options"
}

func (b *Options) Imports(mapper generator.Struct) map[string]string {
	return map[string]string{}
}

func (b *Options) Generate(mapper generator.Struct) ([]byte, error) {
	for _, field := range mapper.Fields {
		if !field.HasTag(RequiredTag) && !field.HasTag(IgnoreTag) {
			fieldName := field.NameOrKindName()
			optionFunc := mapper.Name + strings.Title(fieldName)
			arg := generator.UncapFirst(fieldName)
			b.Printf("func %s(%s %s) func(*%s) {\n", optionFunc, arg, field.Kind.String(), mapper.Name)
			b.Printf("	return func(t *%s) {\n", mapper.Name)
			b.Printf("		t.%s = %s\n", fieldName, arg)
			b.Printf("	}\n")
			b.Printf("}\n\n")
		}
	}

	args := &generator.Scribler{}
	for _, field := range mapper.Fields {
		if field.HasTag(RequiredTag) {
			args.Printf("%s %s,", generator.UncapFirst(field.NameOrKindName()), field.Kind.String())
		}
	}

	structName := mapper.Name
	b.Printf("\nfunc New%sOptions(%s options ...func(*%s)) *%s {\n", structName, args, structName, structName)
	b.Printf("	t := &%s {\n", structName)
	for _, field := range mapper.Fields {
		if field.HasTag(RequiredTag) {
			fieldName := field.NameOrKindName()
			b.Printf("	%s: %s,\n", fieldName, generator.UncapFirst(fieldName))
		}
	}
	b.Printf("	}\n")
	b.Printf("	for _, option := range options {\n")
	b.Printf("		option(t)\n")
	b.Printf("	}\n")
	b.Printf("	return t\n")
	b.Printf("}\n")

	return b.Flush(), nil
}
