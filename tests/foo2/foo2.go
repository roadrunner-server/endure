package foo2

import "github.com/spiral/cascade/tests/foo4"

type DB struct {
}

type S2 struct {
}

func (s2 *S2) Init(db foo4.DB) error {
	println("hello from S2 --> Init")
	return nil
}

func (s2 *S2) Provides() []interface{} {
	return []interface{}{s2.CreateDB}
}

func (s2 *S2) CreateDB() (DB, error) {
	println("hello from S2 --> CreateDB")
	return DB{}, nil
}
