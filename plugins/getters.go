package plugins

import (
	"strings"

	"github.con/quintans/gog/generator"
)

func init() {
	generator.Register(&Getters{})
}

const IgnoreTag = "@ignore"

type Getters struct {
	generator.Scribler
}

func (b *Getters) Name() string {
	return "getter"
}

func (b *Getters) Imports(mapper generator.Struct) map[string]string {
	return map[string]string{}
}

func (b *Getters) Generate(mapper generator.Struct) []byte {
	for _, field := range mapper.Fields {
		if !field.HasTag(IgnoreTag) {
			fieldName := field.NameOrKindName()
			b.Printf("\nfunc (t %s) %s() %s {\n", mapper.Name, strings.Title(fieldName), field.Kind.String())
			b.Printf("  return t.%s\n", fieldName)
			b.Printf("}\n")
		}
	}

	return b.Flush()
}
