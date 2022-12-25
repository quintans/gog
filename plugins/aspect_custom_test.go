package plugins

import (
	"fmt"
	"strings"
	"testing"

	"github.com/quintans/gog/config"
	"github.com/quintans/gog/generator"
)

const (
	AspectTxTag      = "@transactional"
	AspectMonitorTag = "@monitor"
	AspectSecuredTag = "@secured"
)

// Example of how to build a custom aspect generator

func TestCustomPlugin(t *testing.T) {
	generator.Register(&Aspect{})

	tests := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "struct_aspect",
			in: `
			package p

			import "context"

			// gog:aspect
			type Foo struct{}
			
			// gog:@monitor {"threshold": 1}
			// gog:@transactional
			func (f Foo) Handle(ctx context.Context, code string) (int, error) {
				fmt.Println("receive code:", code)
				return 1, nil
			}

			// gog:@secured {"roles": ["user"]}
			func (f Foo) WhoAmI(ctx context.Context) (string, error) {
				// gets user from context, whatever
				return "myname", nil
			}

			// Dummy demonstrates without aspect
			func (f Foo) Dummy(ctx context.Context) int {
				return 1000
			}

			// private demonstrates unexported methods are ignore
			func (f Foo) private(ctx context.Context) int {
				return 1000
			}
			`,
			out: fmt.Sprintf(`// Code generated by gog; DO NOT EDIT.
// Version: %s
package p

import (
	"context"
	"fmt"
	"time"
)

// Generated by gog:aspect

type FooAspect struct {
	Next Foo
}

func (a *FooAspect) Handle(ctx context.Context, code string) (int, error) {
	// transactional aspect
	f1 := func(ctx context.Context, code string) (int, error) {
		var a0 int
		var a1 error
		a1 = fake.WithTx(ctx, func(s string) error {
			a0, a1 = a.Next.Handle(ctx, code)
			return a1
		}) // end of WithTx
		return a0, a1
	} // end of tx

	// monitor aspect
	f0 := func(ctx context.Context, code string) (int, error) {
		now := time.Now()
		defer func() {
			if time.Since(now) > 1*time.Second {
				fmt.Println("slow call")
			}
		}()
		return f1(ctx, code)
	}

	return f0(ctx, code)
}

func (a *FooAspect) WhoAmI(ctx context.Context) (string, error) {
	// secured aspect
	f0 := func(ctx context.Context) (string, error) {
		if err := checkSecurity(ctx, "user"); err != nil {
			return "", err
		}
		a.Next.WhoAmI(ctx)
	}

	return f0(ctx)
}

func (a *FooAspect) Dummy(ctx context.Context) int {
	return a.Next.Dummy(ctx)
}
`, config.Version),
		},
		{
			name: "interface_aspect",
			in: `
			package p

			import "context"

			// gog:aspect
			type Bar interface{
				// gog:@transactional
				Handle(ctx context.Context, code string) (int, error)
			}			
			`,
			out: fmt.Sprintf(`// Code generated by gog; DO NOT EDIT.
// Version: %s
package p

import "context"

// Generated by gog:aspect

type BarAspect struct {
	Next Bar
}

func (a *BarAspect) Handle(ctx context.Context, code string) (int, error) {
	// transactional aspect
	f0 := func(ctx context.Context, code string) (int, error) {
		var a0 int
		var a1 error
		a1 = fake.WithTx(ctx, func(s string) error {
			a0, a1 = a.Next.Handle(ctx, code)
			return a1
		}) // end of WithTx
		return a0, a1
	} // end of tx

	return f0(ctx, code)
}
`, config.Version),
		},
	}
	for _, tt := range tests {
		//! delete me
		if tt.name != "interface_aspect" {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			run(t, tt.in, tt.out)
		})
	}
}

type AspectOptions struct{}

type Aspect struct {
	generator.Scribler
}

func (a Aspect) Name() string {
	return "aspect"
}

func (Aspect) Accepts() []generator.MapperType {
	return []generator.MapperType{
		generator.StructMapper,
		generator.InterfaceMapper,
	}
}

func (a Aspect) Imports(mapper generator.Mapper) map[string]string {
	return map[string]string{}
}

func (a *Aspect) GenerateBody(mapper generator.Mapper) error {
	return a.WriteBody(mapper, AspectOptions{})
}

func (a *Aspect) WriteBody(mapper generator.Mapper, _ AspectOptions) error {
	sName := mapper.GetName() + "Aspect"
	a.BPrintf("type %sAspect struct {\n", mapper.GetName())
	a.BPrintf("Next %s\n", mapper.GetName())
	a.BPrintf("}\n\n")

	for _, m := range mapper.GetMethods() {
		if !m.IsExported() {
			continue
		}

		a.BPrint("func (a *", sName, ") ", m.Signature(true), " {\n")

		methodName := "a.Next." + m.Name()

		tags := m.Tags.Filter(AspectTxTag, AspectMonitorTag, AspectSecuredTag)

		for k := len(tags) - 1; k >= 0; k-- {
			var (
				err  error
				body string
			)

			tag := m.Tags[k]
			switch tag.Name {
			case AspectMonitorTag:
				options := AspectMonitorOptions{}
				if err = tag.Unmarshal(&options); err != nil {
					return err
				}
				body = monitor(&m, methodName, options)
			case AspectTxTag:
				body, err = transactional(&m, methodName)
				if err != nil {
					return err
				}
			case AspectSecuredTag:
				options := AspectSecuredOptions{}
				if err = tag.Unmarshal(&options); err != nil {
					return err
				}
				body, err = secured(&m, methodName, options)
				if err != nil {
					return err
				}
			default:
				continue
			}
			a.BPrintln("// ", tag.Name[1:], " aspect")
			methodName = fmt.Sprintf("f%d", k)
			a.BPrint(methodName, " := ", body, "\n")
		}

		call := fmt.Sprint("(", m.Parameters(true), ")")
		if len(tags) > 0 {
			a.BPrintln("return f0", call)
		} else {
			a.BPrintln("return ", methodName, call)
		}
		a.BPrint("}\n\n")
	}

	return nil
}

type AspectMonitorOptions struct {
	Threshold int
}

type AspectSecuredOptions struct {
	Roles []string
}

func monitor(m *generator.Method, methodName string, options AspectMonitorOptions) string {
	sign := m.Signature(false)
	s := generator.Scribler{}
	s.BPrintf(`func%s{
		now := time.Now()
		defer func(){
			if time.Since(now) > %d*time.Second {
				fmt.Println("slow call")
			}
		}()
	`, sign, options.Threshold)
	if m.HasResults() {
		s.BPrintf("return ")
	}
	s.BPrintln(methodName, "(", m.Parameters(true), ")")
	s.BPrintln("}")

	return s.String()
}

func secured(m *generator.Method, methodName string, options AspectSecuredOptions) (string, error) {
	ctxName := m.ContextArgName()
	if ctxName == "" {
		return "", fmt.Errorf("method %s must have a context.Context argument type to use the 'secured' aspect", m.Name())
	}
	sign := m.Signature(false)
	s := generator.Scribler{}
	s.BPrintf(`func%s{
		if err := checkSecurity(%s, %s); err != nil {
			return %s
		}
		`, sign, ctxName, generator.JoinAround(options.Roles, "\"", "\"", ", "), m.ReturnZerosWithError("err"))

	s.BPrintln(methodName, "(", m.Parameters(true), ")")
	s.BPrintln("}")

	return s.String(), nil
}

func transactional(m *generator.Method, methodName string) (string, error) {
	sign := m.Signature(false)
	s := generator.Scribler{}
	s.BPrintf("func%s{\n", sign)

	var errVar string
	rets := make([]string, 0, len(m.Results))

	for k, a := range m.Results {
		v := fmt.Sprintf("a%d", k)
		s.BPrintf("var %s %s\n", v, a.Kind)
		rets = append(rets, v)
		if a.IsError() {
			errVar = v
		}
	}

	if errVar == "" {
		return "", fmt.Errorf("method %s must return an error type to use the 'transaction' aspect", m.Name())
	}

	s.BPrintln(errVar, " = fake.WithTx(ctx, func(s string) error {")
	s.BPrintln(strings.Join(rets, ","), " = ", methodName, "(", m.Parameters(true), ")")
	if errVar != "" {
		s.BPrintln("return ", errVar)
	} else {
		s.BPrintln("return nil")
	}
	s.BPrintln("}) // end of WithTx") // ends WithTx

	s.BPrintln("return ", strings.Join(rets, ","))
	s.BPrintln("} // end of tx")

	return s.String(), nil
}
