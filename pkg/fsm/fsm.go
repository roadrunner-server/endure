package fsm

import (
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/roadrunner-server/errors"
)

// FSM represents endure finite state machine
type FSM interface {
	// CurrentState returns the current state of the FSM
	CurrentState() State
	// InitialState sets the initial FSM state
	InitialState(st State)
	// Transition used to move from one state to another
	Transition(event Event, args ...interface{}) (interface{}, error)
}

// NewFSM returns new FSM implementation based on initial state and callbacks to move from one state to another
func NewFSM(initialState State, callbacks map[Event]reflect.Method) *FSMImpl {
	st := uint32(initialState)
	return &FSMImpl{
		callbacks:    callbacks,
		currentState: &st,
	}
}

// FSMImpl is endure FSM interface implementation
type FSMImpl struct {
	mutex        sync.Mutex
	currentState *uint32
	callbacks    map[Event]reflect.Method
}

// Event represents FSM event
type Event uint32

func (ev Event) String() string {
	switch ev {
	case Initialize:
		return "Initialize"
	case Start:
		return "Start"
	case Stop:
		return "Stop"
	default:
		return "Unknown event"
	}
}

const (
	// Initialize func
	Initialize Event = iota // Init func
	// Start func
	Start // Serve func
	// Stop func
	Stop // Stop func
)

// State is the FSM current state
type State uint32

func (st State) String() string {
	switch st {
	case Uninitialized:
		return "Uninitialized"
	case Initializing:
		return "Initializing"
	case Initialized:
		return "Initialized"
	case Starting:
		return "Starting"
	case Started:
		return "Started"
	case Stopping:
		return "Stopping"
	case Stopped:
		return "Stopped"
	case Error:
		return "Error"
	default:
		return "Unknown state"
	}
}

const (
	// Uninitialized state
	Uninitialized State = iota
	// Initializing state
	Initializing
	// Initialized state
	Initialized
	// Starting state
	Starting
	// Started state
	Started
	// Stopping state
	Stopping
	// Stopped state
	Stopped
	// Error state
	Error
)

// Acceptors (also called detectors or recognizers) produce binary output,
// indicating whether or not the received input is accepted.
// Each event of an acceptor is either accepting or non accepting.
func (f *FSMImpl) recognizer(event Event) error {
	const op = errors.Op("fsm_recognizer")
	switch event {
	case Initialize:
		if f.current() == Uninitialized || f.current() == Error {
			return nil
		}
		return errors.E(op, errors.Errorf("can't transition from state: %s by event %s", f.current(), event))
	case Start:
		if f.current() != Initialized {
			return errors.E(op, errors.Errorf("can't transition from state: %s by event %s", f.current(), event))
		}
	case Stop:
		if f.current() == Started || f.current() == Error {
			return nil
		}
		return errors.E(op, errors.Errorf("can't transition from state: %s by event %s", f.current(), event))
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

// InitialState (see interface)
func (f *FSMImpl) InitialState(st State) {
	f.set(st)
}

// CurrentState (see interface)
func (f *FSMImpl) CurrentState() State {
	return f.current()
}

/*
Transition moves endure from one state to another
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
	f.mutex.Lock()
	defer f.mutex.Unlock()
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
	default:
		return nil, errors.E("can't be here")
	}
}
