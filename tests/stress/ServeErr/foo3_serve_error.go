package ServeErr

type S3ServeError struct {
}

func (s3 *S3ServeError) Collects() []interface{} {
	return []interface{}{
		s3.SomeOtherDep,
	}
}

func (s3 *S3ServeError) SomeOtherDep(svc *S4ServeError, svc2 *S2) error {
	return nil
}

// Collects on S3
func (s3 *S3ServeError) Init(svc *S2) error {
	return nil
}

func (s3 *S3ServeError) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s3 *S3ServeError) Stop() error {
	return nil
}
