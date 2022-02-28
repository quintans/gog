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

func (b *RequiredArgsConstructor) Imports(mapper generator.Struct) map[string]string {
	return map[string]string{}
}

func (s *RequiredArgsConstructor) GenerateBody(mapper generator.Struct) error {
	args := &generator.Scribler{}
	for _, field := range mapper.Fields {
		if field.HasTag(RequiredTag) {
			args.BPrintf("%s %s,", generator.UncapFirst(field.NameOrKindName()), field.Kind.String())
		}
	}

	structName := mapper.Name
	s.BPrintf("\nfunc New%sRequired(%s) %s {\n", structName, args, structName)
	s.BPrintf(" return %s{\n", structName)
	for _, field := range mapper.Fields {
		if field.HasTag(RequiredTag) {
			fieldName := field.NameOrKindName()
			s.BPrintf("	%s: %s,\n", fieldName, generator.UncapFirst(fieldName))
		}
	}
	s.BPrintf("  }\n")
	s.BPrintf("}\n")

	return nil
}
