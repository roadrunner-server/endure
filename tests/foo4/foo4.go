package foo4

type S4 struct {

}

type DB struct {

}


// No deps
func (s *S4) Init() error {
	println("hello from S4 --> Init")
	return nil
}

// But provide some
func (s *S4) Provides() []interface{} {
	return []interface{}{
		s.CreateAnotherDb,
	}
}

// this is the same type but different packages
func (s *S4) CreateAnotherDb() (DB, error) {
	println("hello from S4 --> CreateAnotherDb")
	return DB{}, nil
}