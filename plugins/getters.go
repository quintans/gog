package plugins

import (
	"fmt"
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

func (*Getters) Accepts() []generator.MapperType {
	return []generator.MapperType{generator.StructMapper}
}

func (b *Getters) Imports(mapper generator.Mapper) map[string]string {
	return map[string]string{}
}

func (b *Getters) GenerateBody(mapper generator.Mapper) error {
	options := GetterOptions{}
	if tag, ok := mapper.GetTags().FindTag(b.Name()); ok {
		if err := tag.Unmarshal(&options); err != nil {
			return err
		}
	}

	err := b.WriteBody(mapper, options)
	if err != nil {
		return fmt.Errorf("writing Getters body: %w", err)
	}
	return nil
}

func (b *Getters) WriteBody(mapper generator.Mapper, options GetterOptions) error {
	var star string
	if options.Pointer {
		star = "*"
	}
	structName := mapper.GetName()
	receiver := generator.UncapFirstSingle(structName)
	for _, field := range mapper.GetFields() {
		if field.HasTag(IgnoreTag) {
			continue
		}
		fieldName := field.NameOrKindName()
		getter := strings.Title(fieldName)
		if field.IsNested() {
			getter = "Get" + getter
		}
		b.BPrintf("\nfunc (%s %s%s) %s() %s {\n", receiver, star, structName, getter, field.Kind.String())
		b.BPrintf("  return %s.%s\n", receiver, fieldName)
		b.BPrintf("}\n")
	}

	return nil
}
