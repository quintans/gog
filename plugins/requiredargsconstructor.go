package plugins

import (
	"github.com/quintans/gog/generator"
)

func init() {
	generator.Register(&RequiredArgsConstructor{})
}

type RequiredArgsConstructor struct {
	generator.Scribler
}

func (b *RequiredArgsConstructor) Name() string {
	return "requiredArgsConstructor"
}

func (b *RequiredArgsConstructor) Imports(mapper *generator.Struct) map[string]string {
	return map[string]string{}
}

func (b *RequiredArgsConstructor) GenerateBody(mapper *generator.Struct) error {
	args := &generator.Scribler{}
	for _, field := range mapper.Fields {
		if field.HasTag(RequiredTag) {
			args.BPrintf("%s %s,", generator.UncapFirst(field.NameOrKindName()), field.Kind.String())
		}
	}

	structName := mapper.Name
	b.BPrintf("\nfunc New%sRequired(%s) %s {\n", structName, args, structName)
	b.BPrintf(" return %s{\n", structName)
	for _, field := range mapper.Fields {
		if field.HasTag(RequiredTag) {
			fieldName := field.NameOrKindName()
			b.BPrintf("	%s: %s,\n", fieldName, generator.UncapFirst(fieldName))
		}
	}
	b.BPrintf("  }\n")
	b.BPrintf("}\n")

	return nil
}
