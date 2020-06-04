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

func (s2 *S2) Init(s4 *test_other_package.S4) {

}

func (s2 *S2) Provides() []interface{} {
	return []interface{}{s2.createDB}
}

func (s2 *S2) createDB() DB {
	return DB{}
}

type S3 struct {
}

func (s3 *S3) Depends() []interface{} {
	return []interface{}{
		s3.SomeOtherDep,
	}
}

func (s3 *S3) SomeOtherDep(svc test_other_package.S4, svc2 S2) error {
	return nil
}

// Depends on S3
func (s3 *S3) Init() {

}


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
	c := NewContainer()

	err := c.Register("s2", &S2{})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Register("s3", &S3{})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Register("s1", &S1{})
	if err != nil {
		t.Fatal(err)
	}

	// this is the same type as S3 create DB
	err = c.Register("s4", &test_other_package.S4{})
	if err != nil {
		t.Fatal(err)
	}

	//err = c.Register("s5", &S5{})
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//err = c.Register("s6", &S6{})
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//err = c.Register("s7", &S7{})
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//err = c.Register("s8", &S8{})
	//if err != nil {
	//	t.Fatal(err)
	//}

	err = c.Init()
	if err != nil {
		t.Fatal(err)
	}
	log.Print(c.servicesGraph.Edges)
}
