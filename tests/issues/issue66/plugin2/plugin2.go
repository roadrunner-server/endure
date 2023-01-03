package plugin2

import (
	"context"
	"fmt"
)

type Plugin2 struct {
}

type IDB3 interface {
	SomeDB3DepMethod()
	Name() string
}

func (p *Plugin2) Init(p3 IDB3) error {
	fmt.Println(p3.Name())
	return nil
}

func (p *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin2) Stop(context.Context) error {
	return nil
}

func (p *Plugin2) SomeDepP2Method() {}
