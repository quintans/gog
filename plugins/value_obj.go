package plugins

import (
	"fmt"
	"strings"

	"github.com/quintans/gog/generator"
)

func init() {
	generator.Register(&ValueObj{})
}

type ValueObj struct {
	generator.Scribler
}

func (b *ValueObj) Name() string {
	return "value"
}

func (*ValueObj) Accepts() []generator.MapperType {
	return []generator.MapperType{generator.StructMapper}
}

func (b *ValueObj) Imports(mapper generator.Mapper) map[string]string {
	return map[string]string{}
}

func (b *ValueObj) GenerateBody(mapper generator.Mapper) error {
	allArgs := &AllArgsConstructor{}
	allArgs.WriteBody(mapper, AllArgsConstructorOptions{})
	b.Body.Write(allArgs.Flush())

	getters := Getters{}
	err := getters.WriteBody(mapper, GetterOptions{})
	if err != nil {
		return fmt.Errorf("writing ValueObj body: %w", err)
	}
	b.BPrintf("\n")
	b.Body.Write(getters.Body.Bytes())

	for _, field := range mapper.GetFields() {
		if field.HasTag(WitherTag) {
			b.genWither(mapper, field)
		}
	}

	_ = PrintIsZero(&b.Scribler, mapper)
	_ = PrintString(&b.Scribler, mapper)

	return nil
}

func (b *ValueObj) genWither(mapper generator.Mapper, field generator.Field) {
	fieldName := field.NameOrKindName()
	receiver := generator.UncapFirstSingle(mapper.GetName())
	wither := "With" + strings.Title(fieldName)
	if _, ok := mapper.FindMethod(wither); !ok {
		b.BPrintf("\nfunc (%s %s) %s(%s %s) %s {\n", receiver, mapper.GetName(), wither, fieldName, field.Kind.String(), mapper.GetName())
		b.BPrintf("  return %s {\n", mapper.GetName())
		for _, f := range mapper.GetFields() {
			fn := f.NameOrKindName()
			if fn == fieldName {
				b.BPrintf("%s: %s,\n", fn, fieldName)
			} else {
				b.BPrintf("%s: %s.%s,\n", fn, receiver, fn)
			}
		}
		b.BPrintf("}\n")
		b.BPrintf("}\n")
	}
}
