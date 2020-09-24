package ServeRetryErr

type FOO2DB struct {
}

type S2 struct {
}

func (s2 *S2) Init() error {
	return nil
}

func (s2 *S2) Close() error {
	return nil
}

func (s2 *S2) Configure() error {
	return nil
}

func (s2 *S2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s2 *S2) Stop() error {
	return nil
}
