package context

import (
	"errors"
	"internal/reflectlite"
	"sync"
	"sync/atomic"
	"time"
)

type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key any) any
}

var Canceled = errors.New("context canceled")

var DeadlineExceeded error = deadlineExceededError{}

//-------------------------------------------------------------------------------------------

type deadlineExceededError struct{}

func (deadlineExceededError) Error() string { return "context deadline exceeded" }

func (deadlineExceededError) Timeout() bool { return true }

func (deadlineExceededError) Temporary() bool { return true }

//-------------------------------------------------------------------------------------------

type emptyCtx struct{}

func (emptyCtx) Deadline() (deadline time.Time, ok bool) { return }

func (emptyCtx) Done() <-chan struct{} { return nil }

func (emptyCtx) Err() error { return nil }

func (emptyCtx) Value(key any) any { return nil }

//-------------------------------------------------------------------------------------------

type backgroundCtx struct { emptyCtx }

func (backgroundCtx) String() string { return "context.Background" }

//-------------------------------------------------------------------------------------------

type todoCtx struct { emptyCtx }

func (todoCtx) String() string { return "context.TODO" }

//-------------------------------------------------------------------------------------------

func Background() Context { return backgroundCtx{} }

//-------------------------------------------------------------------------------------------

func TODO() Context {
	return todoCtx{}
}

//-------------------------------------------------------------------------------------------

type CancelFunc func()

func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
	c := withCancel(parent)
	return c, func() { c.cancel(true, Canceled, nil) }
}

//-------------------------------------------------------------------------------------------

type CancelCauseFunc func(cause error)

func WithCancelCause(parent Context) (ctx Context, cancel CancelCauseFunc) {
	c := withCancel(parent)
	return c, func(cause error) { c.cancel(true, Canceled, cause) }
}

//-------------------------------------------------------------------------------------------

func withCancel(parent Context) *cancelCtx {
	if parent == nil {
		panic("cannot create context from nil parent")
	}
	c := &cancelCtx{}
	c.propagateCancel(parent, c)
	return c
}
























































