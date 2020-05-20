package cascade

import (
	"log"
	"testing"
)

type DB struct {
}

type S1 struct {
}

func (s1 *S1) Init(s2 S2, db DB) {
}

type S2 struct {
}

func (s2 *S2) Init(s3 S3) {

}

type S3 struct {
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

	c.Init()
	log.Print(c.services.dependecies)
}
