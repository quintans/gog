package plugins

import (
	"strings"

	"github.com/quintans/gog/generator"
)

const WitherTag = "@wither"

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
	return map[string]string{
		"\"github.com/quintans/gog\"": "",
	}
}

func (b *ValueObj) Generate(mapper generator.Struct) []byte {
	allArgs := &AllArgsConstructor{}
	b.Body.Write(allArgs.Generate(mapper))

	for _, field := range mapper.Fields {
		if !field.HasTag(IgnoreTag) {
			b.genGetter(mapper, field)
		}
		if field.HasTag(WitherTag) {
			b.genWither(mapper, field)
		}
	}

	return b.Flush()
}

func (b *ValueObj) genGetter(mapper generator.Struct, field generator.Field) {
	fieldName := field.NameOrKindName()
	if _, ok := mapper.FindMethod(strings.Title(fieldName)); !ok {
		b.Printf("\nfunc (t %s) %s() %s {\n", mapper.Name, strings.Title(fieldName), field.Kind.String())
		b.Printf("  return t.%s\n", fieldName)
		b.Printf("}\n")
	}
}

func (b *ValueObj) genWither(mapper generator.Struct, field generator.Field) {
	fieldName := field.NameOrKindName()
	wither := "With" + strings.Title(fieldName)
	if _, ok := mapper.FindMethod(wither); !ok {
		b.Printf("\nfunc (t %s) %s(%s %s) %s {\n", mapper.Name, wither, fieldName, field.Kind.String(), mapper.Name)
		b.Printf("  return %s {\n", mapper.Name)
		for _, f := range mapper.Fields {
			fn := f.NameOrKindName()
			if fn == fieldName {
				b.Printf("%s: %s,\n", fn, fieldName)
			} else {
				b.Printf("%s: t.%s,\n", fn, fn)
			}
		}
		b.Printf("}\n")
		b.Printf("}\n")
	}
}
