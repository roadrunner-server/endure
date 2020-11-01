package endure

import (
	"reflect"
	"sync/atomic"

	//"github.com/spiral/endure/structures"
	"github.com/spiral/errors"
)

type FSM interface {
	CurrentState() State
	InitialState(st State)
	Transition(event Event, args ...interface{}) (interface{}, error)
}

func NewFSM(initialState State, callbacks map[Event]reflect.Method) FSM {
	// callbacks is the pairs EVENT -> Func to invoke
	st := uint32(initialState)
	return &FSMImpl{
		callbacks:    callbacks,
		currentState: &st,
	}
}

type FSMImpl struct {
	currentState *uint32
	callbacks    map[Event]reflect.Method
}

type Event uint32

const (
	Initialize Event = iota // Init func
	Start                   // Serve func
	Stop                    // Stop func
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

// Acceptors (also called detectors or recognizers) produce binary output,
// indicating whether or not the received input is accepted.
// Each event of an acceptor is either accepting or non accepting.
func (f *FSMImpl) recognizer(event Event) error {
	switch event {
	case Initialize:
		if f.current() != Uninitialized {
			return errors.E("wrong state")
		}
	case Start:
		if f.current() != Initialized {
			return errors.E("wrong state")
		}
	case Stop:
		if f.current() != Started {
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

func (f *FSMImpl) InitialState(st State) {
	f.set(st)
}

func (f *FSMImpl) CurrentState() State {
	return f.current()
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
func (f *FSMImpl) Transition(event Event, args ...interface{}) (interface{}, error) {
	err := f.recognizer(event)
	if err != nil {
		return nil, err
	}

	switch event {
	case Initialize:
		f.set(Initializing)
		method := f.callbacks[event]
		values := make([]reflect.Value, 0, len(args))
		for _, v := range args {
			values = append(values, reflect.ValueOf(v))
		}

		ret := method.Func.Call(values)
		if ret[0].Interface() != nil {
			if ret[0].Interface().(error) != nil {
				f.set(Error)
				return nil, ret[0].Interface().(error)
			}
		}

		f.set(Initialized)
		return nil, nil
	case Start:
		f.set(Starting)
		method := f.callbacks[event]
		values := make([]reflect.Value, 0, len(args))
		for _, v := range args {
			values = append(values, reflect.ValueOf(v))
		}

		ret := method.Func.Call(values)
		if ret[1].Interface() != nil {
			if ret[1].Interface().(error) != nil {
				f.set(Error)
				return nil, ret[1].Interface().(error)
			}
		}

		f.set(Started)
		return ret[0].Interface(), nil
	//run Serve
	case Stop:
		f.set(Stopping)
		method := f.callbacks[event]
		values := make([]reflect.Value, 0, len(args))
		for _, v := range args {
			values = append(values, reflect.ValueOf(v))
		}

		ret := method.Func.Call(values)
		if ret[0].Interface() != nil {
			if ret[0].Interface().(error) != nil {
				f.set(Error)
				return nil, ret[0].Interface().(error)
			}
		}

		f.set(Stopped)
		return nil, nil
	//run internal_stop
	default:
		return nil, errors.E("can't be here")
	}
}
