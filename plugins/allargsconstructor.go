package plugins

import (
	"github.com/quintans/gog/generator"
)

func init() {
	generator.Register(&AllArgsConstructor{})
}

type AllArgsConstructor struct {
	generator.Scribler
}

func (b *AllArgsConstructor) Name() string {
	return "allArgsConstructor"
}

func (b *AllArgsConstructor) Imports(mapper generator.Struct) map[string]string {
	return map[string]string{}
}

func (s *AllArgsConstructor) Generate(mapper generator.Struct) []byte {
	structName := mapper.Name

	args := &generator.Scribler{}
	for _, field := range mapper.Fields {
		args.Printf("%s %s,", generator.UncapFirst(field.NameOrKindName()), field.Kind.String())
	}

	s.Printf("\nfunc New%s(%s) %s {\n", structName, args, structName)
	s.Printf(" return %s{\n", structName)
	for _, field := range mapper.Fields {
		fieldName := field.NameOrKindName()
		s.Printf("	%s: %s,\n", fieldName, generator.UncapFirst(fieldName))
	}
	s.Printf("  }\n")
	s.Printf("}\n")

	return s.Flush()
}
