package foo2

import "github.com/spiral/cascade/tests/foo4"

type DB struct {
}

type S2 struct {
}

func (s2 *S2) Init(db *foo4.DB) error {
	println("hello from S2 --> Init")
	println("S4 in S2: " + db.Name + "And changing name to --> S4 greeting you, teacher")
	return nil
}

func (s2 *S2) Provides() []interface{} {
	return []interface{}{s2.CreateDB}
}

func (s2 *S2) CreateDB() (DB, error) {
	return DB{}, nil
}
