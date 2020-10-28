package InitErr

type S2Err struct {
}

func (s2 *S2Err) Init() error {
	return nil
}

func (s2 *S2Err) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s2 *S2Err) Stop() error {
	return nil
}
