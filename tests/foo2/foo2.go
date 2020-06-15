package foo2

import "github.com/spiral/cascade/tests/foo4"

type DB struct {
}

type S2 struct {
}

func (s2 *S2) Init(db *foo4.DB) error {
	println("hello from S2 --> Init")
	println("S4 in S2: " + db.Name + ", and changing the name to the --> S4 greeting you, teacher")
	db.Name = "S4 greeting you, teacher"
	return nil
}

func (s2 *S2) Provides() []interface{} {
	return []interface{}{s2.CreateDB}
}

func (s2 *S2) CreateDB() (DB, error) {
	println("hello from S2 --> CreateDB")
	return DB{}, nil
}

func (s2 *S2) Serve(upstream chan interface{}) error {
	return nil
}

func (s2 *S2) Stop() error {
	return nil
}