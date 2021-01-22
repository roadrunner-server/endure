package interfaces

import (
	"testing"
	"time"

	"github.com/spiral/endure/pkg/container"
	"github.com/spiral/endure/tests/interfaces/named/randominterface"
	"github.com/spiral/endure/tests/interfaces/named/registers"
	"github.com/spiral/endure/tests/interfaces/named/registersfail"
	"github.com/spiral/endure/tests/interfaces/plugins/plugin1"
	"github.com/spiral/endure/tests/interfaces/plugins/plugin10"
	"github.com/spiral/endure/tests/interfaces/plugins/plugin2"
	"github.com/spiral/endure/tests/interfaces/plugins/plugin3"
	"github.com/spiral/endure/tests/interfaces/plugins/plugin4"
	"github.com/spiral/endure/tests/interfaces/plugins/plugin5"
	"github.com/spiral/endure/tests/interfaces/plugins/plugin6"
	"github.com/spiral/endure/tests/interfaces/plugins/plugin7"
	"github.com/spiral/endure/tests/interfaces/plugins/plugin8"
	"github.com/spiral/endure/tests/interfaces/plugins/plugin9"
	notImplPlugin1 "github.com/spiral/endure/tests/interfaces/service/not_implemented_service/plugin1"
	notImplPlugin2 "github.com/spiral/endure/tests/interfaces/service/not_implemented_service/plugin2"
	"github.com/stretchr/testify/assert"
)

func TestEndure_Interfaces_OK(t *testing.T) {
	c, err := endure.NewContainer(nil)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&plugin1.Plugin1{}))
	assert.NoError(t, c.Register(&plugin2.Plugin2{}))
	err = c.Init()
	if err != nil {
		t.Fatal(err)
	}

	res, err := c.Serve()
	assert.NoError(t, err)

	go func() {
		for r := range res {
			if r.Error != nil {
				assert.NoError(t, r.Error)
				return
			}
		}
	}()

	time.Sleep(time.Second * 2)

	assert.NoError(t, c.Stop())
	time.Sleep(time.Second * 1)
}

func TestEndure_InterfacesCollects_Ok(t *testing.T) {
	c, err := endure.NewContainer(nil)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&plugin3.Plugin3{}))
	assert.NoError(t, c.Register(&plugin4.Plugin4{}))
	assert.NoError(t, c.Register(&plugin5.Plugin5{}))

	assert.NoError(t, c.Init())

	_, err = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}

func TestEndure_NamedProvides_Ok(t *testing.T) {
	c, err := endure.NewContainer(nil)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&registers.Plugin2{}))
	assert.NoError(t, c.Register(&registers.Plugin1{}))

	assert.NoError(t, c.Init())

	_, err = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}

func TestEndure_NamedProvides_NotImplement_Ok(t *testing.T) {
	c, err := endure.NewContainer(nil)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&randominterface.Plugin2{}))
	assert.NoError(t, c.Register(&randominterface.Plugin1{}))

	assert.NoError(t, c.Init())

	_, err = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}

func TestEndure_NamedProvides_WrongType_Fail(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			println("test should panic")
		}
	}()
	c, err := endure.NewContainer(nil)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&registersfail.Plugin2{}))
	assert.NoError(t, c.Register(&registersfail.Plugin1{}))

	assert.Error(t, c.Init())

	_, err = c.Serve()
	assert.Error(t, err)

	assert.NoError(t, c.Stop())
}

func TestEndure_ServiceInterface_NotImplemented_Ok(t *testing.T) {
	c, err := endure.NewContainer(nil)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&notImplPlugin1.Foo{}))
	assert.NoError(t, c.Register(&notImplPlugin2.Foo{}))

	assert.NoError(t, c.Init())

	_, err = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}

func Test_MultiplyProvidesSameInterface(t *testing.T) {
	c, err := endure.NewContainer(nil)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&plugin6.Plugin{}))
	assert.NoError(t, c.Register(&plugin6.Plugin2{}))
	assert.NoError(t, c.Register(&plugin6.Plugin3{}))
	err = c.Init()
	if err != nil {
		t.Fatal(err)
	}

	res, err := c.Serve()
	assert.NoError(t, err)

	go func() {
		for r := range res {
			if r.Error != nil {
				assert.NoError(t, r.Error)
				return
			}
		}
	}()

	time.Sleep(time.Second * 2)

	assert.NoError(t, c.Stop())
	time.Sleep(time.Second * 1)
}

func Test_MultiplyCollectsInterface(t *testing.T) {
	c, err := endure.NewContainer(nil)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&plugin7.Plugin7{}))
	assert.NoError(t, c.Register(&plugin8.Plugin8{}))
	assert.NoError(t, c.Register(&plugin9.Plugin9{}))
	assert.NoError(t, c.Register(&plugin10.Plugin10{}))
	err = c.Init()
	if err != nil {
		t.Fatal(err)
	}

	res, err := c.Serve()
	assert.NoError(t, err)

	go func() {
		for r := range res {
			if r.Error != nil {
				assert.NoError(t, r.Error)
				return
			}
		}
	}()

	time.Sleep(time.Second * 2)

	assert.NoError(t, c.Stop())
	time.Sleep(time.Second * 1)
}
