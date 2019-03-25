package logger

import (
	"context"
	"io"
	"os"

	"github.com/blend/go-sdk/async"
)

const (

	// DefaultListenerName is a default.
	DefaultListenerName = "default"

	// DefaultRecoverPanics is a default.
	DefaultRecoverPanics = true
)

// New returns a new logger with a given set of enabled flags, without a writer provisioned.
func New(options ...Option) *Logger {
	l := &Logger{
		Latch:         async.NewLatch(),
		Formatter:     NewTextFormatter(),
		Output:        NewInterlockedWriter(os.Stdout),
		RecoverPanics: DefaultRecoverPanics,
		Flags:         NewFlags(),
	}
	l.Context = NewContext(l)
	for _, option := range options {
		option(l)
	}
	return l
}

// Logger is a handler for various logging events with descendent handlers.
type Logger struct {
	*async.Latch
	*Flags
	*Context

	RecoverPanics bool

	Output    io.Writer
	Formatter WriteFormatter
	Errors    chan error
	Listeners map[string]map[string]*Worker
}

// HasListeners returns if there are registered listener for an event.
func (l *Logger) HasListeners(flag string) bool {
	l.Lock()
	defer l.Unlock()

	if l.Listeners == nil {
		return false
	}
	listeners, ok := l.Listeners[flag]
	if !ok {
		return false
	}
	return len(listeners) > 0
}

// HasListener returns if a specific listener is registerd for a flag.
func (l *Logger) HasListener(flag, listenerName string) bool {
	l.Lock()
	defer l.Unlock()

	if l.Listeners == nil {
		return false
	}
	workers, ok := l.Listeners[flag]
	if !ok {
		return false
	}
	_, ok = workers[listenerName]
	return ok
}

// Listen adds a listener for a given flag.
func (l *Logger) Listen(flag, listenerName string, listener Listener) {
	l.Lock()
	defer l.Unlock()

	if l.Listeners == nil {
		l.Listeners = make(map[string]map[string]*Worker)
	}

	w := NewWorker(listener)
	if listeners, ok := l.Listeners[flag]; ok {
		listeners[listenerName] = w
	} else {
		l.Listeners[flag] = map[string]*Worker{
			listenerName: w,
		}
	}
	go w.Start()
	<-w.NotifyStarted()
}

// RemoveListeners clears *all* listeners for a Flag.
func (l *Logger) RemoveListeners(flag string) {
	l.Lock()
	defer l.Unlock()

	if l.Listeners == nil {
		return
	}

	listeners, ok := l.Listeners[flag]
	if !ok {
		return
	}

	for _, l := range listeners {
		l.Stop()
	}

	delete(l.Listeners, flag)
}

// RemoveListener clears a specific listener for a Flag.
func (l *Logger) RemoveListener(flag, listenerName string) {
	l.Lock()
	defer l.Unlock()

	if l.Listeners == nil {
		return
	}

	listeners, ok := l.Listeners[flag]
	if !ok {
		return
	}

	worker, ok := listeners[listenerName]
	if !ok {
		return
	}

	worker.Stop()
	<-worker.NotifyStopped()

	delete(listeners, listenerName)
	if len(listeners) == 0 {
		delete(l.Listeners, flag)
	}
}

// Trigger fires the listeners for a given event asynchronously.
// The invocations will be queued in a work queue per listener.
// There are no order guarantees on when these events will be processed across listeners.
// This call will not block on the event listeners, but will block on writing the event to the formatted output.
func (l *Logger) Trigger(ctx context.Context, e Event) {
	flag := e.Flag()
	if !l.IsEnabled(flag) {
		return
	}

	if typed, isTyped := e.(EnabledProvider); isTyped && !typed.IsEnabled() {
		return
	}

	var listeners map[string]*Worker
	l.Lock()
	if l.Listeners != nil {
		if flagListeners, ok := l.Listeners[flag]; ok {
			listeners = flagListeners
		}
	}
	l.Unlock()

	for _, listener := range listeners {
		listener.Work <- e
	}
}

// Write writes an event synchronously to the writer either as a normal even or as an error.
func (l *Logger) Write(ctx context.Context, e Event) {
	// check if the event controls if it should be written or not.
	if typed, isTyped := e.(WritableProvider); isTyped && !typed.IsWritable() {
		return
	}

	if l.Formatter != nil && l.Output != nil {
		if err := l.Formatter.WriteFormat(ctx, l.Output, e); err != nil && l.Errors != nil {
			l.Errors <- err
		}
	}
}

// --------------------------------------------------------------------------------
// finalizers
// --------------------------------------------------------------------------------

// Close releases shared resources for the agent.
func (l *Logger) Close() error {
	l.Stopping()

	if l.Flags != nil {
		l.Flags.SetNone()
	}

	for _, listeners := range l.Listeners {
		for _, listener := range listeners {
			listener.Stop()
			<-listener.NotifyStopped()
		}
	}
	for key := range l.Listeners {
		delete(l.Listeners, key)
	}
	l.Listeners = nil

	l.Stopped()
	return nil
}

// Drain waits for the agent to finish its queue of events before closing.
func (l *Logger) Drain() error {
	for _, workers := range l.Listeners {
		for _, worker := range workers {
			worker.Drain()
		}
	}
	return nil
}
