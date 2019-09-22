package example

// Dto1 is the base for the generating the builder
//
//gog:allArgsConstructor
//gog:builder
//gog:getter
type Dto1 struct {
	age func(string) int
	// Name is fine
	//gog:@required
	name  string
	value int64
	sex   bool
	other *Dto2
}

func (dto *Dto1) setValue(value int64) error {
	dto.value = value
	return nil
}

// Greet is ignored
func Greet(s string) string {
	return "hello " + s
}

// Dto2 for a second builder
//
//gog:builder
//gog:getter
type Dto2 struct {
	cenas []int
}
