package endure

import (
	"sync/atomic"

	//"github.com/spiral/endure/structures"
	"github.com/spiral/errors"
)

type FSM interface {
	Initial()
	Transition(event Event) error
}

type FSMImpl struct {
	container    *Endure
	currentState *uint32
}

func NewFSM(e *Endure) FSM {
	tmp := uint32(0)
	return &FSMImpl{
		currentState: &tmp,
		container:    e,
	}
}

func (f *FSMImpl) Initial() {

}

type Event uint32

const (
	Initialize Event = iota
	Start
	Stop
)

type State uint32

const (
	Uninitialized State = iota
	Initializing
	Initialized
	Starting
	Started
	Stopping
	Stopped
	Error // ??
)

type state struct{}

// Acceptors (also called detectors or recognizers) produce binary output,
// indicating whether or not the received input is accepted.
// Each event of an acceptor is either accepting or non accepting.
func (f *FSMImpl) recognizer(event Event) error {
	switch event {
	case Initialize:
		if f.current() > Uninitialized {
			return errors.E("wrong state")
		}
	case Start:
		if f.current() > Initialized {
			return errors.E("wrong state")
		}
	case Stop:
		if f.current() > Started {
			return errors.E("wrong state")
		}
	}

	return nil
}

// SetState sets state
func (f *FSMImpl) set(st State) {
	atomic.StoreUint32(f.currentState, uint32(st))
}

// CurrentState returns current state
func (f *FSMImpl) current() State {
	return State(atomic.LoadUint32(f.currentState))
}

/*
Rules:
Transition table:
Event -> Init. Error on other events (Start, Stop)
1. Initializing -> Initialized
Event -> Start. Error on other events (Initialize, Stop)
2. Starting -> Started
Event -> Stop. Error on other events (Start, Initialize)
3. Stopping -> Stopped
*/
func (f *FSMImpl) Transition(event Event) error {
	err := f.recognizer(event)
	if err != nil {
		return err
	}

	switch event {
	case Initialize:
		f.set(Initializing)
		err := f.container.Init()
		// run Init
		if err != nil {
			return err
		}
		f.set(Initialized)
		return nil
	case Start:
	//run Serve
	case Stop:
	//run stop
	default:
		return errors.E("can't be here")
	}
	return nil
}
