package plugins

import (
	"strings"

	"github.com/quintans/gog/generator"
)

func init() {
	generator.Register(&Getters{})
}

type GetterOptions struct {
	Pointer bool
}

type Getters struct {
	generator.Scribler
}

func (b *Getters) Name() string {
	return "getters"
}

func (b *Getters) Imports(mapper generator.Struct) map[string]string {
	return map[string]string{}
}

func (b *Getters) GenerateBody(mapper generator.Struct) error {
	options := GetterOptions{}
	if tag, ok := mapper.FindTag(b.Name()); ok {
		if err := tag.Unmarshal(&options); err != nil {
			return err
		}
	}

	b.WriteBody(mapper, options)
	return nil
}

func (b *Getters) WriteBody(mapper generator.Struct, options GetterOptions) error {
	var star string
	if options.Pointer {
		star = "*"
	}
	structName := mapper.Name
	receiver := generator.UncapFirstSingle(structName)
	for _, field := range mapper.Fields {
		if !field.HasTag(IgnoreTag) {
			fieldName := field.NameOrKindName()
			getter := strings.Title(fieldName)
			if field.IsNested() {
				getter = "Get" + getter
			}
			b.BPrintf("\nfunc (%s %s%s) %s() %s {\n", receiver, star, structName, getter, field.Kind.String())
			b.BPrintf("  return %s.%s\n", receiver, fieldName)
			b.BPrintf("}\n")
		}
	}

	return nil
}
