package main

func init() {
	Register(&AllArgsConstructor{})
}

type AllArgsConstructor struct {
	Scribler
}

func (b *AllArgsConstructor) Name() string {
	return "allArgsConstructor"
}

func (b *AllArgsConstructor) Imports(mapper Struct) map[string]string {
	return map[string]string{}
}

func (b *AllArgsConstructor) Generate(mapper Struct) []byte {
	structName := mapper.Name

	var args Scribler
	for _, field := range mapper.Fields {
		if field.HasTag(requiredTag) {
			args.Printf("%s %s,", UncapFirst(field.NameOrKindName()), field.Kind.String())
		}
	}

	b.Printf("\nfunc New%s(%s) %s {\n", structName, args.Body.String(), structName)
	b.Printf(" return %s{\n", structName)
	for _, field := range mapper.Fields {
		if field.HasTag(requiredTag) {
			fieldName := field.NameOrKindName()
			b.Printf("	%s: %s,\n", fieldName, UncapFirst(fieldName))
		}
	}
	b.Printf("  }\n")
	b.Printf("}\n")

	return b.Flush()
}
