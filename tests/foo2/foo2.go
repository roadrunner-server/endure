package foo2

import "github.com/spiral/cascade/tests/foo4"

type DB struct {
}

type S2 struct {
}

func (s2 *S2) Init(db foo4.DB) {

}

func (s2 *S2) Provides() []interface{} {
	return []interface{}{s2.createDB}
}

func (s2 *S2) createDB() DB {
	return DB{}
}
