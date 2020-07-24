package foo2

import (
	"errors"
	"math/rand"
	"time"

	"github.com/spiral/endure/tests/foo4"
)

type DBServeErr struct {
}

type S2ServeErr struct {
}

func (s2 *S2ServeErr) Init(db *foo4.DB) error {
	s := rand.Intn(10)
	// just random
	if s == 5 {
		return errors.New("random error during init from S3")
	}
	return nil
}

func (s2 *S2ServeErr) Provides() []interface{} {
	return []interface{}{s2.CreateDB}
}

func (s2 *S2ServeErr) CreateDB() (DB, error) {
	return DB{}, nil
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
