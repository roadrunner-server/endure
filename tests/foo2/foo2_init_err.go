package foo2

import (
	"errors"

	"github.com/spiral/cascade/tests/foo4"
)

type DBErr struct {
}

type S2Err struct {
}

func (s2 *S2Err) Init(db *foo4.DB) error {
	println("hello from S2Err --> Init")
	return errors.New("init test error")
}

func (s2 *S2Err) Provides() []interface{} {
	return []interface{}{s2.CreateDB}
}

func (s2 *S2Err) CreateDB() (DB, error) {
	println("hello from S2Err --> CreateDB")
	return DB{}, nil
}

func (s2 *S2Err) Close() error {
	return nil
}

func (s2 *S2Err) Configure() chan error {
	errCh := make(chan error, 1)
	println("S2Err: configuring")
	return errCh
}

func (s2 *S2Err) Serve() chan error {
	errCh := make(chan error, 1)
	println("S2Err: serving")
	return errCh
}

func (s2 *S2Err) Stop() error {
	println("S2Err: stopping")
	return nil
}