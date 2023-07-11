package main

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	domain      = "www.example.com"
	baseURL     = "https://" + domain
	cardURL     = baseURL + "/api/public/some/%s/path/%s"
	queryParams = "?page=1&size=10000"
)

var (
	ready = false

	cntAdded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mysvc",
			Name:      "cnt_added",
			Help:      "Total items added",
		},
		[]string{"caller"},
	)
)

// SvcOptions is the representation of the options availble to the FooBar service
type SvcOptions struct {
	FooOption *string
}

// FooBar is a generic service
type FooBar struct {
	Context   context.Context
	Logger    log.Logger
	FooOption string
}

// NewFooBarSvc creates an instance of the FooBar Service.
func NewFooBarSvc(ctx context.Context, l log.Logger, o *SvcOptions) *FooBar {
	return &FooBar{
		Context:   ctx,
		Logger:    l,
		FooOption: *o.FooOption,
	}
}

// IsReady returns a bool describing the state of the service.
// Output:
//
//	True when the service is processing SQS messages
//	Otherwise False
func (svc *FooBar) IsReady() bool {
	return ready
}

// type RequestDelay func()
func requestDelay(delay time.Duration) func() {
	return func() {
		time.Sleep(delay * time.Millisecond)
	}
}

type ctxReqDelay struct{}

func contextWithRequestDelayFn(ctx context.Context, reqDelay func()) context.Context {
	return context.WithValue(ctx, ctxReqDelay{}, reqDelay)
}

func requestDelayFnFromContext(ctx context.Context) func() {
	if d, ok := ctx.Value(ctxReqDelay{}).(func()); ok {
		return d
	}

	panic("fail to retrieve request delay func")
}

// Start is the main business logic loop.
func (svc *FooBar) Start() error {
	level.Info(svc.Logger).Log("msg", "starting service")

	// service is ready
	ready = true
	level.Info(svc.Logger).Log("msg", "service ready")

	// Main service loop.
	for ready {
		// do something
	}

	level.Info(svc.Logger).Log("msg", "service task completed")

	return nil
}

// Stop instructs the service to stop processing new messages.
func (svc *FooBar) Stop() {
	level.Info(svc.Logger).Log("msg", "stopping service")
	ready = false
}
