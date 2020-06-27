package foo5

type Reader interface {
	WRead() // just stupid name
}


type S5 struct {
}

func (s *S5) WRead() {
	println("S5: WReading...")
}

type DB struct {
	Name string
}

// No deps
func (s *S5) Init() error {
	println("hello from S5 --> Init")
	return nil
}

func (s *S5) Configure() error {
	println("S5: configuring")
	return nil
}

func (s *S5) Serve() chan error {
	errCh := make(chan error, 1)
	println("S5: serving")
	return errCh
}

func (s *S5) Close() error {
	println("S5: closing")
	return nil
}

func (s *S5) Stop() error {
	println("S5: stopping")
	return nil
}