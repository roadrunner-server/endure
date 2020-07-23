package mainthread

import (
	"errors"
	"time"
)

var number int = 0

type Foo struct {
}

func (f *Foo) Init() error {
	number += 1
	return nil
}

func (f *Foo) Configure() error {
	if number > 1 {
		return errors.New("test error when num > 1")
	}
	return nil
}

func (f *Foo) Close() error {
	return nil
}

func (f *Foo) Serve() chan error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- errors.New("test error in serve")
	}()
	return errCh
}

func (f *Foo) Stop() error {
	return nil
}

/////////////////////////////////////////////////

var number2 int = 0

type Foo2 struct {
}

func (f *Foo2) Init() error {
	if number2 > 0 {
		return errors.New("test error when num > 1")
	}
	return nil
}

func (f *Foo2) Configure() error {
	return nil
}

func (f *Foo2) Close() error {
	return nil
}

func (f *Foo2) Serve() chan error {
	errCh := make(chan error, 1)
	number2 += 1
	go func() {
		errCh <- errors.New("test error2")
	}()
	return errCh
}

func (f *Foo2) Stop() error {
	return nil
}

/////////////////////////////////////////////////

var number3 int = 0

type Foo3 struct {
}

func (f *Foo3) Init() error {
	return nil
}

func (f *Foo3) Configure() error {
	return nil
}

func (f *Foo3) Close() error {
	return nil
}

func (f *Foo3) Serve() chan error {
	errCh := make(chan error, 1)
	if number3 == 0 {
		number3++
		go func() {
			time.Sleep(time.Second * 3)
			errCh <- errors.New("test error3")
		}()
	} else {
		return nil
	}
	return errCh
}

func (f *Foo3) Stop() error {
	return nil
}
