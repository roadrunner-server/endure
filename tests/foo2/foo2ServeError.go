package foo2

import (
	"errors"
	"time"

	"github.com/spiral/cascade/tests/foo4"
)

type DBServeErr struct {
}

type S2ServeErr struct {
}

func (s2 *S2ServeErr) Init(db *foo4.DB) error {
	println("hello from S2ServeErr --> Init")
	println("S4 in S2ServeErr: " + db.Name + ", and changing the name to the --> S4 greeting you, teacher")
	db.Name = "S4 greeting you, teacher"
	return nil
}

func (s2 *S2ServeErr) Provides() []interface{} {
	return []interface{}{s2.CreateDB}
}

func (s2 *S2ServeErr) CreateDB() (DB, error) {
	println("hello from S2ServeErr --> CreateDB")
	return DB{}, nil
}

func (s2 *S2ServeErr) Close() error {
	return nil
}

func (s2 *S2ServeErr) Configure() error {
	println("S2ServeErr: configuring")
	return nil
}

func (s2 *S2ServeErr) Serve() chan error {
	errCh := make(chan error, 1)
	go func() {
		time.Sleep(time.Second * 5)
		errCh <- errors.New("S2ServeErr test err in serve")
	}()
	println("S2ServeErr: serving")
	return errCh
}

func (s2 *S2ServeErr) Stop() error {
	println("S2ServeErr: stopping")
	return nil
}