package plugin1

import (
	"github.com/spiral/endure/tests/happy_scenarios/plugin2"
	"github.com/spiral/endure/tests/happy_scenarios/plugin4"
)

type S1 struct {
}

func (s1 *S1) Depends() []interface{} {
	return []interface{}{
		s1.AddService,
	}
}

func (s1 *S1) AddService(svc *plugin4.DB) error {
	return nil
}

// Depends on S2 and DB (S3 in the current case)
func (s1 *S1) Init(s2 *plugin2.S2, db *plugin2.DB) error {
	return nil
}

func (s1 *S1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s1 *S1) Stop() error {
	return nil
}
