package foo1

type S1One struct {
}

func (s1 *S1One) Init() error {
	return nil
}

func (s1 *S1One) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s1 *S1One) Configure() error {
	return nil
}

func (s1 *S1One) Close() error {
	return nil
}

func (s1 *S1One) Stop() error {
	return nil
}
