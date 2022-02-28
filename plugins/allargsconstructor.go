package plugins

import (
	"github.com/quintans/gog/generator"
)

func init() {
	generator.Register(&AllArgsConstructor{})
}

type AllArgsConstructorOptions struct{}

type AllArgsConstructor struct {
	generator.Scribler
}

func (s AllArgsConstructor) Name() string {
	return "allArgsConstructor"
}

func (s AllArgsConstructor) Imports(mapper generator.Struct) map[string]string {
	return map[string]string{}
}

func (s *AllArgsConstructor) GenerateBody(mapper generator.Struct) error {
	s.WriteBody(mapper, AllArgsConstructorOptions{})
	return nil
}

func (c *AllArgsConstructor) WriteBody(mapper generator.Struct, _ AllArgsConstructorOptions) {
	args := &generator.Scribler{}
	hasError := false
	for _, field := range mapper.Fields {
		args.BPrintf("%s %s,\n", generator.UncapFirst(field.NameOrKindName()), field.Kind.String())
		if !hasError && field.HasTag(RequiredTag) {
			hasError = true
		}
	}
	structName := mapper.Name
	receiver := generator.UncapFirstSingle(structName)
	s := &generator.Scribler{}
	if hasError {
		_ = PrintZeroCheck(s, mapper, "")
	}

	s.BPrintf("%s := %s{\n", receiver, structName)
	for _, field := range mapper.Fields {
		fieldName := field.NameOrKindName()
		s.BPrintf("	%s: %s,\n", fieldName, field.NameForField())
	}
	s.BPrintf("  }\n")

	hasError = PrintValidate(s, mapper, receiver) || hasError

	retCode := structName
	if hasError {
		retCode = "(" + retCode + ", error)"
	}
	c.BPrintf("\nfunc New%s(\n%s) %s {\n", structName, args, retCode)
	c.BPrintf("%s\n", s)
	c.BPrintf("return %s", receiver)
	if hasError {
		c.BPrintf(", nil")
	}
	c.BPrintf("\n}\n")
}
