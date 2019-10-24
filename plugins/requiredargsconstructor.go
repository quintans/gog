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

func (s *RequiredArgsConstructor) Generate(mapper generator.Struct) ([]byte, error) {
	args := &generator.Scribler{}
	for _, field := range mapper.Fields {
		if field.HasTag(RequiredTag) {
			args.Printf("%s %s,", generator.UncapFirst(field.NameOrKindName()), field.Kind.String())
		}
	}

	structName := mapper.Name
	s.Printf("\nfunc New%sRequired(%s) %s {\n", structName, args, structName)
	s.Printf(" return %s{\n", structName)
	for _, field := range mapper.Fields {
		if field.HasTag(RequiredTag) {
			fieldName := field.NameOrKindName()
			s.Printf("	%s: %s,\n", fieldName, generator.UncapFirst(fieldName))
		}
	}
	s.Printf("  }\n")
	s.Printf("}\n")

	return s.Flush(), nil
}
