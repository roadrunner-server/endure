package ServeErr

type DB struct {
}

type S2 struct {
}

func (s2 *S2) Init(db *FOO4DB) error {
	return nil
}

func (s2 *S2) Provides() []interface{} {
	return []interface{}{s2.CreateDB}
}

func (s2 *S2) CreateDB() (DB, error) {
	return DB{}, nil
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
