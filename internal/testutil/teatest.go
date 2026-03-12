package testutil

import (
	"bytes"
	"io"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
)

// TestModel is a model being tested, wrapping a tea.Program for v2.
type TestModel struct {
	program *tea.Program
	out     io.ReadWriter
	doneCh  chan bool
	done    sync.Once
}

// TestOption configures test model creation.
type TestOption func(*testModelOpts)

type testModelOpts struct {
	width  int
	height int
}

// WithInitialTermSize sets the initial terminal size.
func WithInitialTermSize(w, h int) TestOption {
	return func(o *testModelOpts) {
		o.width = w
		o.height = h
	}
}

// NewTestModel creates a test model for integration testing with Bubble Tea v2.
func NewTestModel(tb testing.TB, m tea.Model, options ...TestOption) *TestModel {
	opts := testModelOpts{}
	for _, o := range options {
		o(&opts)
	}

	tm := &TestModel{
		out:    safe(bytes.NewBuffer(nil)),
		doneCh: make(chan bool, 1),
	}

	progOpts := []tea.ProgramOption{
		tea.WithInput(nil),
		tea.WithOutput(tm.out),
		tea.WithoutSignals(),
		tea.WithColorProfile(colorprofile.Ascii),
	}
	if opts.width > 0 && opts.height > 0 {
		progOpts = append(progOpts, tea.WithWindowSize(opts.width, opts.height))
	}

	tm.program = tea.NewProgram(m, progOpts...)

	go func() {
		if _, err := tm.program.Run(); err != nil {
			tb.Logf("program exited with error: %s", err)
		}
		tm.doneCh <- true
	}()

	return tm
}

// Send sends a message to the program.
func (tm *TestModel) Send(msg tea.Msg) {
	tm.program.Send(msg)
}

// Output returns the program's output reader.
func (tm *TestModel) Output() io.Reader {
	return tm.out
}

// WaitFinished waits for the program to exit.
func (tm *TestModel) WaitFinished(tb testing.TB, opts ...FinalOpt) {
	tm.done.Do(func() {
		fopts := finalOpts{}
		for _, o := range opts {
			o(&fopts)
		}
		if fopts.timeout > 0 {
			select {
			case <-time.After(fopts.timeout):
				tb.Fatalf("timeout after %s", fopts.timeout)
			case <-tm.doneCh:
			}
		} else {
			<-tm.doneCh
		}
	})
}

// FinalOpt configures WaitFinished behavior.
type FinalOpt func(*finalOpts)

type finalOpts struct {
	timeout time.Duration
}

// WithFinalTimeout sets the timeout for WaitFinished.
func WithFinalTimeout(d time.Duration) FinalOpt {
	return func(o *finalOpts) {
		o.timeout = d
	}
}

// WaitForOption configures WaitFor behavior.
type WaitForOption func(*waitForOpts)

type waitForOpts struct {
	duration      time.Duration
	checkInterval time.Duration
}

// WithDuration sets the maximum wait duration.
func WithDuration(d time.Duration) WaitForOption {
	return func(o *waitForOpts) {
		o.duration = d
	}
}

// WaitFor waits until the condition is met on the output.
func WaitFor(tb testing.TB, r io.Reader, condition func([]byte) bool, options ...WaitForOption) {
	tb.Helper()
	opts := waitForOpts{
		duration:      time.Second,
		checkInterval: 50 * time.Millisecond,
	}
	for _, o := range options {
		o(&opts)
	}

	var b bytes.Buffer
	start := time.Now()
	for time.Since(start) <= opts.duration {
		if _, err := io.ReadAll(io.TeeReader(r, &b)); err != nil {
			tb.Fatalf("WaitFor: %v", err)
		}
		if condition(b.Bytes()) {
			return
		}
		time.Sleep(opts.checkInterval)
	}
	tb.Fatalf("WaitFor: condition not met after %s. Last output:\n%s", opts.duration, b.String())
}

func safe(rw io.ReadWriter) io.ReadWriter {
	return &safeReadWriter{rw: rw}
}

type safeReadWriter struct {
	rw io.ReadWriter
	m  sync.RWMutex
}

func (s *safeReadWriter) Read(p []byte) (int, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.rw.Read(p)
}

func (s *safeReadWriter) Write(p []byte) (int, error) {
	s.m.Lock()
	defer s.m.Unlock()
	return s.rw.Write(p)
}

