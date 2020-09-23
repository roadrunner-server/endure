package ServeRetryErr

import (
	"errors"
	"math/rand"
	"time"
)

type DBServeErr struct {
}

type S2ServeErr struct {
}

func (s2 *S2ServeErr) Init(db *FOO4DB) error {
	s := rand.Intn(10)
	// just random
	if s == 5 {
		return errors.New("random error during init from S2")
	}
	return nil
}

func (s2 *S2ServeErr) Provides() []interface{} {
	return []interface{}{s2.CreateDB}
}

func (s2 *S2ServeErr) CreateDB() (FOO4DB, error) {
	return FOO4DB{}, nil
}

func (s2 *S2ServeErr) Close() error {
	return nil
}

func (s2 *S2ServeErr) Configure() error {
	return nil
}

func (s2 *S2ServeErr) Serve() chan error {
	errCh := make(chan error, 1)
	go func() {
		time.Sleep(time.Second * 1)
		errCh <- errors.New("test error in S2ServeErr")
	}()
	return errCh
}

func (s2 *S2ServeErr) Stop() error {
	return nil
}
