package issues

import (
	"fmt"
	"testing"
	"time"

	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/endure/tests/issues/issue33"
	issue55_p1 "github.com/roadrunner-server/endure/tests/issues/issue55/plugin1"
	issue55_p2 "github.com/roadrunner-server/endure/tests/issues/issue55/plugin2"
	issue55_p3 "github.com/roadrunner-server/endure/tests/issues/issue55/plugin3"

	issue66_p1 "github.com/roadrunner-server/endure/tests/issues/issue66/plugin1"
	issue66_p2 "github.com/roadrunner-server/endure/tests/issues/issue66/plugin2"
	issue66_p3 "github.com/roadrunner-server/endure/tests/issues/issue66/plugin3"

	issue54_p1 "github.com/roadrunner-server/endure/tests/issues/issue54/plugin1"
	issue54_p2 "github.com/roadrunner-server/endure/tests/issues/issue54/plugin2"
	issue54_p3 "github.com/roadrunner-server/endure/tests/issues/issue54/plugin3"

	issue84_struct_p1 "github.com/roadrunner-server/endure/tests/issues/issue84/structs/plugin1"
	issue84_struct_p2 "github.com/roadrunner-server/endure/tests/issues/issue84/structs/plugin2"
	issue84_struct_p3 "github.com/roadrunner-server/endure/tests/issues/issue84/structs/plugin3"

	issue84_interface_p1 "github.com/roadrunner-server/endure/tests/issues/issue84/interfaces/plugin1"
	issue84_interface_p2 "github.com/roadrunner-server/endure/tests/issues/issue84/interfaces/plugin2"
	issue84_interface_p3 "github.com/roadrunner-server/endure/tests/issues/issue84/interfaces/plugin3"

	issue84_interfaces_structs_p1 "github.com/roadrunner-server/endure/tests/issues/issue84/interfaces_structs/plugin1"
	issue84_interfaces_structs_p2 "github.com/roadrunner-server/endure/tests/issues/issue84/interfaces_structs/plugin2"
	issue84_interfaces_structs_p3 "github.com/roadrunner-server/endure/tests/issues/issue84/interfaces_structs/plugin3"

	issue84_one_alive_p1 "github.com/roadrunner-server/endure/tests/issues/issue84/one_alive/plugin1"
	issue84_one_alive_p2 "github.com/roadrunner-server/endure/tests/issues/issue84/one_alive/plugin2"
	issue84_one_alive_p3 "github.com/roadrunner-server/endure/tests/issues/issue84/one_alive/plugin3"
	"github.com/stretchr/testify/assert"
)

// Provided structure instead of function
func TestEndure_Issue33(t *testing.T) {
	c, err := endure.NewContainer(nil)
	assert.NoError(t, err)

	assert.Error(t, c.Register(&issue33.Plugin1{}))
	assert.NoError(t, c.Register(&issue33.Plugin2{}))
}

// https://github.com/roadrunner-server/endure/issues/55
// Plugin2 froze execution
// Call Stop on the container
// Should be only 1 stop
func TestEndure_Issue55(t *testing.T) {
	container, err := endure.NewContainer(nil)
	assert.NoError(t, err)

	assert.NoError(t, container.Register(&issue55_p1.Plugin1{}))
	assert.NoError(t, container.Register(&issue55_p2.Plugin2{}))
	assert.NoError(t, container.Register(&issue55_p3.Plugin3{}))

	assert.NoError(t, container.Init())

	stopCh := make(chan struct{}, 2)

	go func() {
		time.Sleep(time.Second)
		stopCh <- struct{}{}
	}()

	resCh, err := container.Serve()
	if err == nil {
		t.Fatal(err)
	}

	for {
		select {
		case e := <-resCh:
			fmt.Println(e)
			// first stop
			stopCh <- struct{}{}
			// at the same moment second stop
			err = container.Stop()
			if err != nil {
				t.Fatal(err)
			}
			return
		case <-stopCh:
			err = container.Stop()
			if err != nil {
				t.Fatal(err)
			}
			return
		}
	}
}

func TestIssue54(t *testing.T) {
	c, err := endure.NewContainer(nil)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&issue54_p1.Plugin1{}))
	assert.NoError(t, c.Register(&issue54_p2.Plugin2{}))
	assert.NoError(t, c.Register(&issue54_p3.Plugin3{}))

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

func TestIssue66(t *testing.T) {
	c, err := endure.NewContainer(nil)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&issue66_p1.Plugin1{}))
	assert.NoError(t, c.Register(&issue66_p2.Plugin2{}))
	assert.NoError(t, c.Register(&issue66_p3.Plugin3{}))

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

func TestIssue84_structs_all_disabled(t *testing.T) {
	c, err := endure.NewContainer(nil)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&issue84_struct_p1.Plugin1{}))
	assert.NoError(t, c.Register(&issue84_struct_p2.Plugin2{}))
	assert.NoError(t, c.Register(&issue84_struct_p3.Plugin3{}))

	err = c.Init()
	assert.Error(t, err)
}

func TestIssue84_interfaces_all_disabled(t *testing.T) {
	c, err := endure.NewContainer(nil)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&issue84_interface_p2.Plugin2{}))
	assert.NoError(t, c.Register(&issue84_interface_p1.Plugin1{}))
	assert.NoError(t, c.Register(&issue84_interface_p3.Plugin3{}))

	err = c.Init()
	assert.Error(t, err)
}

func TestIssue84_structs_interface_all_disabled_interface(t *testing.T) {
	c, err := endure.NewContainer(nil)
	assert.NoError(t, err)
	assert.NoError(t, c.Register(&issue84_interfaces_structs_p2.Plugin2{}))
	assert.NoError(t, c.Register(&issue84_interfaces_structs_p1.Plugin1{}))
	assert.NoError(t, c.Register(&issue84_interfaces_structs_p3.Plugin3{}))

	err = c.Init()
	assert.Error(t, err)
}

func TestIssue84_one_alive(t *testing.T) {
	c, err := endure.NewContainer(nil)
	assert.NoError(t, err)
	assert.NoError(t, c.Register(&issue84_one_alive_p1.Plugin1{}))
	assert.NoError(t, c.Register(&issue84_one_alive_p2.Plugin2{}))
	assert.NoError(t, c.Register(&issue84_one_alive_p3.Plugin3{}))

	err = c.Init()
	assert.NoError(t, err)
}
