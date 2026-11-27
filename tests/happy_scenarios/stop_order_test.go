package happy_scenarios

import (
	"context"
	stderrors "errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/endure/v2/dep"
	"github.com/roadrunner-server/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	stopLogIface     interface{ logMarker() }
	stopDBIface      interface{ dbMarker() }
	stopOtelIface    interface{ otelMarker() }
	stopBridgeIface  interface{ bridgeMarker() }
	stopDualA        interface{ dualA() }
	stopDualB        interface{ dualB() }
	stopDualCollects interface{ dualCollects() }
	stopCycleB       interface{ cycleB() }
	stopCycleC       interface{ cycleC() }
)

type stopState struct {
	mu           sync.Mutex
	order        []string
	stoppedCount map[string]int
}

func newStopState() *stopState {
	return &stopState{
		stoppedCount: make(map[string]int),
	}
}

func (s *stopState) record(event string) {
	s.mu.Lock()
	s.order = append(s.order, event)
	s.mu.Unlock()
}

func (s *stopState) markStopped(name string) {
	s.mu.Lock()
	s.stoppedCount[name]++
	s.order = append(s.order, name+":end")
	s.mu.Unlock()
}

func (s *stopState) indexOf(event string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, e := range s.order {
		if e == event {
			return i
		}
	}
	return -1
}

func (s *stopState) reset() {
	s.mu.Lock()
	s.order = s.order[:0]
	clear(s.stoppedCount)
	s.mu.Unlock()
}

func assertStoppedBefore(t *testing.T, state *stopState, first, second, msg string) {
	t.Helper()
	firstIdx := state.indexOf(first)
	secondIdx := state.indexOf(second)
	require.NotEqual(t, -1, firstIdx, "missing event %q", first)
	require.NotEqual(t, -1, secondIdx, "missing event %q", second)
	assert.Less(t, firstIdx, secondIdx, msg)
}

type stopLoggerService struct {
	state   *stopState
	stopErr error
}

func (p *stopLoggerService) Init() error       { return nil }
func (p *stopLoggerService) Serve() chan error { return make(chan error, 1) }
func (p *stopLoggerService) logMarker()        {}
func (p *stopLoggerService) Provides() []*dep.Out {
	return []*dep.Out{
		dep.Bind((*stopLogIface)(nil), p.ProvideLog),
	}
}
func (p *stopLoggerService) ProvideLog() *stopLoggerService { return p }
func (p *stopLoggerService) Stop(_ context.Context) error {
	p.state.record("logger:start")
	p.state.markStopped("logger")
	return p.stopErr
}

type stopDBService struct {
	state *stopState
	delay time.Duration
}

func (p *stopDBService) Init(_ stopLogIface) error { return nil }
func (p *stopDBService) Serve() chan error         { return make(chan error, 1) }
func (p *stopDBService) dbMarker()                 {}
func (p *stopDBService) Stop(_ context.Context) error {
	p.state.record("db:start")
	if p.delay > 0 {
		time.Sleep(p.delay)
	}
	p.state.markStopped("db")
	return nil
}

type stopHTTPService struct {
	state *stopState
	delay time.Duration
}

func (p *stopHTTPService) Init(_ stopLogIface, _ stopDBIface) error { return nil }
func (p *stopHTTPService) Serve() chan error                        { return make(chan error, 1) }
func (p *stopHTTPService) Stop(_ context.Context) error {
	p.state.record("http:start")
	if p.delay > 0 {
		time.Sleep(p.delay)
	}
	p.state.markStopped("http")
	return nil
}

type stopHTTPLogOnlyService struct {
	state       *stopState
	delay       time.Duration
	entered     chan struct{}
	waitForPeer chan struct{}
	stopErr     error
}

func (p *stopHTTPLogOnlyService) Init(_ stopLogIface) error { return nil }
func (p *stopHTTPLogOnlyService) Serve() chan error         { return make(chan error, 1) }
func (p *stopHTTPLogOnlyService) Stop(ctx context.Context) error {
	p.state.record("http:start")
	if p.entered != nil {
		close(p.entered)
	}
	if p.waitForPeer != nil {
		select {
		case <-p.waitForPeer:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if p.delay > 0 {
		time.Sleep(p.delay)
	}
	p.state.markStopped("http")
	return p.stopErr
}

type stopJobsService struct {
	state       *stopState
	entered     chan struct{}
	waitForPeer chan struct{}
}

func (p *stopJobsService) Init(_ stopLogIface) error { return nil }
func (p *stopJobsService) Serve() chan error         { return make(chan error, 1) }
func (p *stopJobsService) Stop(ctx context.Context) error {
	p.state.record("jobs:start")
	if p.entered != nil {
		close(p.entered)
	}
	if p.waitForPeer != nil {
		select {
		case <-p.waitForPeer:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	p.state.markStopped("jobs")
	return nil
}

type stopNonServiceBridge struct{}

func (p *stopNonServiceBridge) Init(_ stopLogIface) error { return nil }
func (p *stopNonServiceBridge) bridgeMarker()             {}

type stopRPCService struct {
	state *stopState
	delay time.Duration
}

func (p *stopRPCService) Init() error       { return nil }
func (p *stopRPCService) Serve() chan error { return make(chan error, 1) }
func (p *stopRPCService) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(_ any) {}, (*stopBridgeIface)(nil)),
	}
}
func (p *stopRPCService) Stop(_ context.Context) error {
	p.state.record("rpc:start")
	if p.delay > 0 {
		time.Sleep(p.delay)
	}
	p.state.markStopped("rpc")
	return nil
}

type stopDualProvider struct {
	state *stopState
}

func (p *stopDualProvider) Init() error       { return nil }
func (p *stopDualProvider) Serve() chan error { return make(chan error, 1) }
func (p *stopDualProvider) dualA()            {}
func (p *stopDualProvider) dualB()            {}
func (p *stopDualProvider) dualCollects()     {}
func (p *stopDualProvider) Stop(_ context.Context) error {
	p.state.record("provider:start")
	p.state.markStopped("provider")
	return nil
}

type stopDualConsumer struct {
	state *stopState
	delay time.Duration
}

func (p *stopDualConsumer) Init(_ stopDualA, _ stopDualB) error { return nil }
func (p *stopDualConsumer) Serve() chan error                   { return make(chan error, 1) }
func (p *stopDualConsumer) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(_ any) {}, (*stopDualCollects)(nil)),
	}
}
func (p *stopDualConsumer) Stop(_ context.Context) error {
	p.state.record("consumer:start")
	if p.delay > 0 {
		time.Sleep(p.delay)
	}
	p.state.markStopped("consumer")
	return nil
}

type stopDisabledDependent struct{}

func (p *stopDisabledDependent) Init(_ stopLogIface) error {
	return errors.E(errors.Disabled)
}
func (p *stopDisabledDependent) Serve() chan error            { return make(chan error, 1) }
func (p *stopDisabledDependent) Stop(_ context.Context) error { return nil }

type stopOtelService struct {
	state *stopState
	delay time.Duration
}

func (p *stopOtelService) Init(_ stopLogIface) error { return nil }
func (p *stopOtelService) Serve() chan error         { return make(chan error, 1) }
func (p *stopOtelService) otelMarker()               {}
func (p *stopOtelService) Stop(_ context.Context) error {
	p.state.record("otel:start")
	if p.delay > 0 {
		time.Sleep(p.delay)
	}
	p.state.markStopped("otel")
	return nil
}

type stopHTTPWithOtelService struct {
	state *stopState
	delay time.Duration
}

func (p *stopHTTPWithOtelService) Init(_ stopLogIface) error { return nil }
func (p *stopHTTPWithOtelService) Serve() chan error         { return make(chan error, 1) }
func (p *stopHTTPWithOtelService) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(_ any) {}, (*stopOtelIface)(nil)),
	}
}
func (p *stopHTTPWithOtelService) Stop(_ context.Context) error {
	p.state.record("http:start")
	if p.delay > 0 {
		time.Sleep(p.delay)
	}
	p.state.markStopped("http")
	return nil
}

type stopCyclicMemberB struct{}

func (p *stopCyclicMemberB) Init(_ stopLogIface, _ stopCycleC) error { return nil }
func (p *stopCyclicMemberB) Serve() chan error                       { return make(chan error, 1) }
func (p *stopCyclicMemberB) Stop(_ context.Context) error            { return nil }
func (p *stopCyclicMemberB) cycleB()                                 {}

type stopCyclicMemberC struct{}

func (p *stopCyclicMemberC) Init(_ stopCycleB) error      { return nil }
func (p *stopCyclicMemberC) Serve() chan error            { return make(chan error, 1) }
func (p *stopCyclicMemberC) Stop(_ context.Context) error { return nil }
func (p *stopCyclicMemberC) cycleC()                      {}

type (
	stopPadIface      interface{ padMarker() }
	stopRaceSinkIface interface{ raceSink() }
)

type stopWeightedLoggerService struct {
	stopLoggerService
}

func (p *stopWeightedLoggerService) Weight() uint { return 100 }

type stopPad[T any] struct{}

func (*stopPad[T]) Init() error { return nil }
func (*stopPad[T]) padMarker()  {}

type stopRaceSink struct{}

func (*stopRaceSink) Init() error { return nil }
func (*stopRaceSink) raceSink()   {}
func (*stopRaceSink) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(_ any) {}, (*stopPadIface)(nil)),
	}
}

type stopRaceBridge struct{}

func (*stopRaceBridge) Init(_ stopLogIface, _ stopRaceSinkIface) error { return nil }

func TestEndure_StopOrder(t *testing.T) {
	t.Parallel()

	errHTTPStop := stderrors.New("http stop failed")
	errLoggerStop := stderrors.New("logger stop failed")

	tests := []struct {
		name  string
		setup func() (*stopState, []any)
		check func(t *testing.T, state *stopState, stopErr error)
	}{
		{
			name: "chain stops in reverse dependency order",
			setup: func() (*stopState, []any) {
				state := newStopState()
				return state, []any{
					&stopLoggerService{state: state},
					&stopDBService{state: state, delay: 40 * time.Millisecond},
					&stopHTTPService{state: state, delay: 40 * time.Millisecond},
				}
			},
			check: func(t *testing.T, state *stopState, stopErr error) {
				require.NoError(t, stopErr)
				assertStoppedBefore(t, state, "http:end", "db:start", "http must finish stopping before db starts")
				assertStoppedBefore(t, state, "db:end", "logger:start", "db must finish stopping before logger starts")
			},
		},
		{
			name: "independent siblings stop in parallel before shared provider",
			setup: func() (*stopState, []any) {
				state := newStopState()
				httpEntered := make(chan struct{})
				jobsEntered := make(chan struct{})
				return state, []any{
					&stopLoggerService{state: state},
					&stopHTTPLogOnlyService{state: state, entered: httpEntered, waitForPeer: jobsEntered},
					&stopJobsService{state: state, entered: jobsEntered, waitForPeer: httpEntered},
				}
			},
			check: func(t *testing.T, state *stopState, stopErr error) {
				require.NoError(t, stopErr)
				assertStoppedBefore(t, state, "http:end", "logger:start", "http must finish stopping before logger starts")
				assertStoppedBefore(t, state, "jobs:end", "logger:start", "jobs must finish stopping before logger starts")
			},
		},
		{
			name: "transitive dependency through non-Service bridge preserves stop order",
			setup: func() (*stopState, []any) {
				state := newStopState()
				return state, []any{
					&stopLoggerService{state: state},
					&stopNonServiceBridge{},
					&stopRPCService{state: state, delay: 40 * time.Millisecond},
				}
			},
			check: func(t *testing.T, state *stopState, stopErr error) {
				require.NoError(t, stopErr)
				assertStoppedBefore(t, state, "rpc:end", "logger:start", "rpc must finish stopping before logger starts")
			},
		},
		{
			name: "duplicate edges between same vertices deduplicate without deadlock",
			setup: func() (*stopState, []any) {
				state := newStopState()
				return state, []any{
					&stopDualProvider{state: state},
					&stopDualConsumer{state: state, delay: 40 * time.Millisecond},
				}
			},
			check: func(t *testing.T, state *stopState, stopErr error) {
				require.NoError(t, stopErr)
				assertStoppedBefore(t, state, "consumer:end", "provider:start", "consumer must finish stopping before provider starts")
				assert.Equal(t, 1, state.stoppedCount["provider"])
				assert.Equal(t, 1, state.stoppedCount["consumer"])
			},
		},
		{
			name: "disabled dependent does not block provider stop",
			setup: func() (*stopState, []any) {
				state := newStopState()
				return state, []any{
					&stopLoggerService{state: state},
					&stopDisabledDependent{},
				}
			},
			check: func(t *testing.T, state *stopState, stopErr error) {
				require.NoError(t, stopErr)
				assert.Equal(t, 1, state.stoppedCount["logger"])
			},
		},
		{
			name: "dependent stop error still stops provider and joins errors",
			setup: func() (*stopState, []any) {
				state := newStopState()
				return state, []any{
					&stopLoggerService{state: state, stopErr: errLoggerStop},
					&stopHTTPLogOnlyService{state: state, delay: 40 * time.Millisecond, stopErr: errHTTPStop},
				}
			},
			check: func(t *testing.T, state *stopState, stopErr error) {
				require.ErrorIs(t, stopErr, errHTTPStop)
				require.ErrorIs(t, stopErr, errLoggerStop)
				assertStoppedBefore(t, state, "http:end", "logger:start", "failing dependent must finish before provider starts")
			},
		},
		{
			name: "collects diamond stops http then otel then logger",
			setup: func() (*stopState, []any) {
				state := newStopState()
				return state, []any{
					&stopLoggerService{state: state},
					&stopOtelService{state: state, delay: 40 * time.Millisecond},
					&stopHTTPWithOtelService{state: state, delay: 40 * time.Millisecond},
				}
			},
			check: func(t *testing.T, state *stopState, stopErr error) {
				require.NoError(t, stopErr)
				assertStoppedBefore(t, state, "http:end", "otel:start", "http must finish stopping before otel starts")
				assertStoppedBefore(t, state, "otel:end", "logger:start", "otel must finish stopping before logger starts")
			},
		},
		{
			name: "cyclic dependents excluded from topological order do not block provider stop",
			setup: func() (*stopState, []any) {
				state := newStopState()
				return state, []any{
					&stopLoggerService{state: state},
					&stopCyclicMemberB{},
					&stopCyclicMemberC{},
				}
			},
			check: func(t *testing.T, state *stopState, stopErr error) {
				require.NoError(t, stopErr)
				assert.Equal(t, 1, state.stoppedCount["logger"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			state, plugins := tt.setup()
			c := endure.New(slog.LevelError, endure.GracefulShutdownTimeout(2*time.Second))
			require.NoError(t, c.RegisterAll(plugins...))
			require.NoError(t, c.Init())

			_, err := c.Serve()
			require.NoError(t, err)

			done := make(chan error, 1)
			go func() {
				done <- c.Stop()
			}()

			select {
			case stopErr := <-done:
				tt.check(t, state, stopErr)
			case <-time.After(5 * time.Second):
				t.Fatal("Stop did not complete within 5s")
			}
		})
	}
}

func TestEndure_StopDoesNotDoubleStopProvider(t *testing.T) {
	t.Parallel()

	const iterations = 200

	tests := []struct {
		name  string
		setup func() (*stopState, []any)
		check func(t *testing.T, state *stopState)
	}{
		{
			name: "zero-delay service chain stops each plugin once",
			setup: func() (*stopState, []any) {
				state := newStopState()
				return state, []any{
					&stopLoggerService{state: state},
					&stopDBService{state: state},
					&stopHTTPService{state: state},
				}
			},
			check: func(t *testing.T, state *stopState) {
				assert.Equal(t, 1, state.stoppedCount["http"])
				assert.Equal(t, 1, state.stoppedCount["db"])
				assert.Equal(t, 1, state.stoppedCount["logger"])
				assertStoppedBefore(t, state, "http:end", "db:start", "http must finish stopping before db starts")
				assertStoppedBefore(t, state, "db:end", "logger:start", "db must finish stopping before logger starts")
			},
		},
		{
			name: "non-service dependent stops provider once",
			setup: func() (*stopState, []any) {
				state := newStopState()
				return state, []any{
					&stopLoggerService{state: state},
					&stopNonServiceBridge{},
				}
			},
			check: func(t *testing.T, state *stopState) {
				assert.Equal(t, 1, state.stoppedCount["logger"])
			},
		},
		{
			name: "zero-delay collects diamond stops each plugin once",
			setup: func() (*stopState, []any) {
				state := newStopState()
				return state, []any{
					&stopLoggerService{state: state},
					&stopOtelService{state: state},
					&stopHTTPWithOtelService{state: state},
				}
			},
			check: func(t *testing.T, state *stopState) {
				assert.Equal(t, 1, state.stoppedCount["http"])
				assert.Equal(t, 1, state.stoppedCount["otel"])
				assert.Equal(t, 1, state.stoppedCount["logger"])
				assertStoppedBefore(t, state, "http:end", "otel:start", "http must finish stopping before otel starts")
				assertStoppedBefore(t, state, "otel:end", "logger:start", "otel must finish stopping before logger starts")
			},
		},
		{
			name: "provider separated from fast dependent by independent plugins stops once",
			setup: func() (*stopState, []any) {
				state := newStopState()
				return state, []any{
					&stopWeightedLoggerService{stopLoggerService: stopLoggerService{state: state}},
					&stopPad[[1]byte]{}, &stopPad[[2]byte]{}, &stopPad[[3]byte]{}, &stopPad[[4]byte]{},
					&stopPad[[5]byte]{}, &stopPad[[6]byte]{}, &stopPad[[7]byte]{}, &stopPad[[8]byte]{},
					&stopPad[[9]byte]{}, &stopPad[[10]byte]{}, &stopPad[[11]byte]{}, &stopPad[[12]byte]{},
					&stopPad[[13]byte]{}, &stopPad[[14]byte]{}, &stopPad[[15]byte]{}, &stopPad[[16]byte]{},
					&stopPad[[17]byte]{}, &stopPad[[18]byte]{}, &stopPad[[19]byte]{}, &stopPad[[20]byte]{},
					&stopPad[[21]byte]{}, &stopPad[[22]byte]{}, &stopPad[[23]byte]{}, &stopPad[[24]byte]{},
					&stopPad[[25]byte]{}, &stopPad[[26]byte]{}, &stopPad[[27]byte]{}, &stopPad[[28]byte]{},
					&stopPad[[29]byte]{}, &stopPad[[30]byte]{}, &stopPad[[31]byte]{}, &stopPad[[32]byte]{},
					&stopRaceSink{},
					&stopRaceBridge{},
				}
			},
			check: func(t *testing.T, state *stopState) {
				assert.Equal(t, 1, state.stoppedCount["logger"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			state, plugins := tt.setup()
			c := endure.New(slog.LevelError, endure.GracefulShutdownTimeout(2*time.Second))
			require.NoError(t, c.RegisterAll(plugins...))
			require.NoError(t, c.Init())

			_, err := c.Serve()
			require.NoError(t, err)

			for range iterations {
				state.reset()
				require.NoError(t, c.Stop())
				tt.check(t, state)
			}
		})
	}
}
