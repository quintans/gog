package plugins

import (
	"strings"

	"github.com/quintans/gog/generator"
)

func init() {
	generator.Register(&Getters{})
}

const IgnoreTag = "@ignore"

type GetterOptions struct {
	Star bool
}

type Getters struct {
	generator.Scribler
}

func (b *Getters) Name() string {
	return "getter"
}

func (b *Getters) Imports(mapper generator.Struct) map[string]string {
	return map[string]string{}
}

func (b *Getters) Generate(mapper generator.Struct) ([]byte, error) {
	options := GetterOptions{}
	if tag, ok := mapper.FindTag(b.Name()); ok {
		if err := tag.Unmarshal(&options); err != nil {
			return nil, err
		}
	}
	var star string
	if options.Star {
		star = "*"
	}
	for _, field := range mapper.Fields {
		if !field.HasTag(IgnoreTag) {
			fieldName := field.NameOrKindName()
			b.Printf("\nfunc (t %s%s) %s() %s {\n", star, mapper.Name, strings.Title(fieldName), field.Kind.String())
			b.Printf("  return t.%s\n", fieldName)
			b.Printf("}\n")
		}
	}

	return b.Flush(), nil
}
