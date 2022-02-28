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

func (b *Options) GenerateBody(mapper generator.Struct) error {
	for _, field := range mapper.Fields {
		if !field.HasTag(RequiredTag) && !field.HasTag(IgnoreTag) {
			fieldName := field.NameOrKindName()
			optionFunc := mapper.Name + strings.Title(fieldName)
			arg := generator.UncapFirst(fieldName)
			b.BPrintf("func %s(%s %s) func(*%s) {\n", optionFunc, arg, field.Kind.String(), mapper.Name)
			b.BPrintf("	return func(t *%s) {\n", mapper.Name)
			b.BPrintf("		t.%s = %s\n", fieldName, arg)
			b.BPrintf("	}\n")
			b.BPrintf("}\n\n")
		}
	}

	args := &generator.Scribler{}
	for _, field := range mapper.Fields {
		if field.HasTag(RequiredTag) {
			args.BPrintf("%s %s,", generator.UncapFirst(field.NameOrKindName()), field.Kind.String())
		}
	}

	structName := mapper.Name
	b.BPrintf("\nfunc New%sOptions(%s options ...func(*%s)) *%s {\n", structName, args, structName, structName)
	b.BPrintf("	t := &%s {\n", structName)
	for _, field := range mapper.Fields {
		if field.HasTag(RequiredTag) {
			fieldName := field.NameOrKindName()
			b.BPrintf("	%s: %s,\n", fieldName, generator.UncapFirst(fieldName))
		}
	}
	b.BPrintf("	}\n")
	b.BPrintf("	for _, option := range options {\n")
	b.BPrintf("		option(t)\n")
	b.BPrintf("	}\n")
	b.BPrintf("	return t\n")
	b.BPrintf("}\n")

	return nil
}
