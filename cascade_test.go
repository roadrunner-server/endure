package cascade

import (
	"log"
	"testing"

	"github.com/spiral/cascade/test_other_package"
)

type DB struct {
}

type S1 struct {
}

func (s1 *S1) Depends() []interface{} {
	return []interface{}{
		s1.AddService,
	}
}

func (s1 *S1) AddService(svc test_other_package.S4) error {
	return nil
}

// Depends on S2 and DB (S3 in the current case)
func (s1 *S1) Init(s2 S2, db DB) {
}

type S2 struct {
}

// Depends on S3
func (s2 *S2) Init(s3 S3) {

}

// No Deps, but provides DB dependency
type S3 struct {
}

func (s3 *S3) Init() {

}

func (s3 *S3) Provides() []interface{} {
	return []interface{}{s3.createDB}
}

func (s3 *S3) createDB() DB {
	return DB{}
}

func TestCascade_Init(t *testing.T) {
	c := NewContainer()

	c.Register("s2", &S2{})
	c.Register("s3", &S3{})
	c.Register("s1", &S1{})
	// this is the same type as S3 create DB
	c.Register("s4", &test_other_package.S4{})

	c.Init()
	log.Print(c.servicesGraph.Edges)
}
