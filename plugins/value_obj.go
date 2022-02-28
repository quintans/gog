package plugins

import (
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

func (b *ValueObj) Imports(mapper generator.Struct) map[string]string {
	return map[string]string{}
}

func (b *ValueObj) GenerateBody(mapper generator.Struct) error {
	allArgs := &AllArgsConstructor{}
	allArgs.WriteBody(mapper, AllArgsConstructorOptions{})
	b.Body.Write(allArgs.Flush())

	getters := Getters{}
	getters.WriteBody(mapper, GetterOptions{})
	b.BPrintf("\n")
	b.Body.Write(getters.Body.Bytes())

	for _, field := range mapper.Fields {
		if field.HasTag(WitherTag) {
			b.genWither(mapper, field)
		}
	}

	_ = PrintIsZero(&b.Scribler, mapper)
	_ = PrintString(&b.Scribler, mapper)

	return nil
}

func (b *ValueObj) genWither(mapper generator.Struct, field generator.Field) {
	fieldName := field.NameOrKindName()
	receiver := generator.UncapFirstSingle(mapper.Name)
	wither := "With" + strings.Title(fieldName)
	if _, ok := mapper.FindMethod(wither); !ok {
		b.BPrintf("\nfunc (%s %s) %s(%s %s) %s {\n", receiver, mapper.Name, wither, fieldName, field.Kind.String(), mapper.Name)
		b.BPrintf("  return %s {\n", mapper.Name)
		for _, f := range mapper.Fields {
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
