package plugins

import (
	"fmt"

	"github.com/quintans/gog/generator"
)

func init() {
	generator.Register(&Record{
		allArgs: &AllArgsConstructor{},
		getters: &Getters{},
	})
}

type Record struct {
	generator.Scribler
	allArgs *AllArgsConstructor
	getters *Getters
}

type RecordOptions struct{}

func (s *Record) Name() string {
	return "record"
}

func (s *Record) Imports(mapper *generator.Struct) map[string]string {
	m := s.allArgs.Imports(mapper)
	generator.MergeMaps(m, s.getters.Imports(mapper))
	return m
}

func (s *Record) GenerateBody(mapper *generator.Struct) error {
	return s.WriteBody(mapper, RecordOptions{})
}

func (s *Record) WriteBody(mapper *generator.Struct, _ RecordOptions) error {
	s.allArgs.WriteBody(mapper, AllArgsConstructorOptions{})
	err := s.getters.WriteBody(mapper, GetterOptions{})
	if err != nil {
		return fmt.Errorf("writing Record body: %w", err)
	}

	s.Body.Write(s.allArgs.Flush())
	s.Body.Write(s.getters.Flush())

	_ = PrintIsZero(&s.Scribler, mapper)

	_ = PrintString(&s.Scribler, mapper)

	return nil
}
