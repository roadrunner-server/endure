package tests

import (
	"testing"

	"github.com/spiral/cascade"
	"github.com/spiral/cascade/tests/foo1"
	"github.com/spiral/cascade/tests/foo2"
	"github.com/spiral/cascade/tests/foo3"
	"github.com/spiral/cascade/tests/foo4"
)

//type S8 struct {
//}
//
//func (s8 *S8) Init(s1 S1) {
//
//}
//
//// provides nothing
//func (s8 *S8) Provides() []interface{} {
//	return []interface{}{}
//}
//
//type S7SomeDep struct {
//}
//
//type S7 struct {
//}
//
//func (s7 *S7) Init(s4 *test_other_package.S4) {
//
//}
//
//func (s7 *S7) Depends() []interface{} {
//	return []interface{}{
//		s7.SomeDep,
//	}
//}
//
//func (s7 *S7) Provides() []interface{} {
//	return []interface{}{s7.provideSomeDep}
//}
//
//func (s7 *S7) SomeDep(svc S1) {
//
//}
//
//func (s7 *S7) provideSomeDep() S7SomeDep {
//	return S7SomeDep{}
//}
//
//type S6 struct {
//}
//
//func (s6 *S6) Provides() []interface{} {
//	return []interface{}{s6.createDB}
//}
//
//func (s6 *S6) createDB() DB {
//	return DB{}
//}
//
//type S5 struct {
//}
//
//func (s5 *S5) Init(sd S7SomeDep) {
//
//}

func TestCascade_Init(t *testing.T) {
	c := cascade.NewContainer()

	err := c.Register("s2", &foo2.S2{})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Register("s3", &foo3.S3{})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Register("s1", &foo1.S1{})
	if err != nil {
		t.Fatal(err)
	}

	// this is the same type as S3 create DB
	err = c.Register("s4", &foo4.S4{})
	if err != nil {
		t.Fatal(err)
	}

	err = c.Init()
	if err != nil {
		t.Fatal(err)
	}
	//log.Print(c.servicesGraph.Edges)
}
