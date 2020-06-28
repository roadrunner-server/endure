package foo2

import (
	"errors"
	"time"

	"github.com/spiral/cascade/tests/foo4"
)

type DBErr struct {
}

type S2Err struct {
}

func (s2 *S2Err) Init(db *foo4.DB) error {
	return errors.New("init test error")
}

func (s2 *S2Err) Provides() []interface{} {
	return []interface{}{s2.CreateDB}
}

func (s2 *S2Err) CreateDB() (DB, error) {
	return DB{}, nil
}

func (s2 *S2Err) Close() error {
	return nil
}

func (s2 *S2Err) Configure() error {
	time.Sleep(time.Second)
	return nil
}

func (s2 *S2Err) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s2 *S2Err) Stop() error {
	time.Sleep(time.Second)
	return nil
}