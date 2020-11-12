package plugin2

import "github.com/spiral/endure/tests/happy_scenarios/plugin4"

type DB struct {
}

type S2 struct {
}

func (s2 *S2) Init(db *plugin4.DB) error {
	return nil
}

func (s2 *S2) Provides() []interface{} {
	return []interface{}{s2.CreateDB}
}

func (s2 *S2) CreateDB() (*DB, error) {
	return &DB{}, nil
}

func (s2 *S2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s2 *S2) Stop() error {
	return nil
}
