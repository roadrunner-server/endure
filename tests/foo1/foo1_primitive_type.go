package foo1

type S1Pr struct {
}

// Depends on S2 and DB (S3 in the current case)
func (s1 *S1Pr) Init(a int) error {
	return nil
}

func (s1 *S1Pr) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s1 *S1Pr) Stop() error {
	return nil
}
