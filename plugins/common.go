package plugins

import "github.com/quintans/gog/generator"

const (
	ValidateMethodName = "validate"
	RequiredTag        = "@required"
	IgnoreTag          = "@ignore"
	WitherTag          = "@wither"
)

func PrintValidate(s *generator.Scribler, mapper generator.Struct, receiver string) bool {
	_, ok := mapper.FindMethod(ValidateMethodName)
	if ok {
		structName := mapper.Name
		s.BPrintf("  if err := %s.validate(); err != nil {", receiver)
		s.BPrintf("    return %s{}, err", structName)
		s.BPrintf("  }\n\n")
	}
	return ok
}

func PrintZeroCheck(s *generator.Scribler, mapper generator.Struct, receiver string) bool {
	if receiver != "" {
		receiver += "."
	}
	structName := mapper.Name
	checked := false
	for _, field := range mapper.Fields {
		if field.HasTag(RequiredTag) {
			checked = true
			s.BPrintf("  if %s {\n", field.Kind.ZeroCondition(receiver+field.NameForField()))
			s.BPrintf("    return %s{}, errors.New(\"%s.%s cannot be empty\")\n", structName, structName, field.Name)
			s.BPrintf("  }\n")
		}
	}
	return checked
}

func PrintIsZero(s *generator.Scribler, mapper generator.Struct) bool {
	structName := mapper.Name
	if _, ok := mapper.FindMethod("IsZero"); ok {
		return false
	}

	receiver := generator.UncapFirstSingle(structName)
	s.BPrintf("\nfunc (%s %s) IsZero() bool {\n", receiver, structName)
	comparable := true
	for _, f := range mapper.Fields {
		_, basic := f.Kind.(generator.Basic)
		if !basic {
			comparable = false
		}
	}
	if comparable {
		s.BPrintf("  return %s == %s{}\n", receiver, structName)
	} else {
		last := len(mapper.Fields) - 1
		s.BPrintf("  return ")
		for k, field := range mapper.Fields {
			s.BPrintf("%s", field.Kind.ZeroCondition(receiver+"."+field.NameForField()))
			if k < last {
				s.BPrintf(" ||\n")
			}
		}
	}
	s.BPrintf("}\n")

	return true
}

func PrintString(s *generator.Scribler, mapper generator.Struct) bool {
	structName := mapper.Name
	if _, ok := mapper.FindMethod("String"); ok {
		return false
	}

	receiver := generator.UncapFirstSingle(structName)
	s.BPrintf("\nfunc (%s %s) String() string {\n", receiver, structName)

	s.BPrintf("  return fmt.Sprintf(\"%s{", structName)
	for k, field := range mapper.Fields {
		if k > 0 {
			s.BPrintf(", ")
		}
		s.BPrintf("%s: %%+v", field.NameOrKindName())
	}
	s.BPrintf("}\", ")
	for k, field := range mapper.Fields {
		if k > 0 {
			s.BPrintf(", ")
		}
		s.BPrintf("%s.%s", receiver, field.NameOrKindName())
	}
	s.BPrintf(")\n")
	s.BPrintf("}\n")

	return true
}
