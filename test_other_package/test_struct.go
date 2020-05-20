package test_other_package

import "database/sql"

type S4 struct {

}

// No deps
func (s *S4) Init() error {
	return nil
}

// But provide some
func (s *S4) Provides() []interface{} {
	return []interface{}{
		s.createAnotherDB,
	}
}

// this is the same type but different packages
func (s *S4) createAnotherDB() *sql.DB {
	return &sql.DB{}
}
