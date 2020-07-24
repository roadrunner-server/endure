package foo2

import (
	"github.com/spiral/endure/tests/foo4"
)

type S2V struct {
}

func (s2 *S2V) Init(db *foo4.DBV) error {
	return nil
}

func (s2 *S2V) Close() error {
	return nil
}

func (s2 *S2V) Configure() error {
	return nil
}

func (s2 *S2V) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s2 *S2V) Stop() error {
	return nil
}
