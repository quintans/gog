package example

// Dto1 is the base for the generating the builder
//
//gog:builder
//gog:getter
type Dto1 struct {
	age func(string) int
	// Name is fine
	//gog:@required
	name string // forward
	// valuable
	value int64
	// ???
	sex   bool
	other *Dto2
} // struct comment

// Dto2 for a second builder
//
//gog:builder
//gog:getter
type Dto2 struct {
	cenas []int
}
