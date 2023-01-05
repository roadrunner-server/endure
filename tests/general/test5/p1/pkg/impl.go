package pkg

type Foo struct {
	N string
}

func InitFoo() *Foo {
	return &Foo{}
}

func (f *Foo) Init(val string) error {
	f.N = val
	return nil
}

func (f *Foo) FooBar() string {
	return f.N
}

func (f *Foo) Name() string {
	return "Foo"
}
