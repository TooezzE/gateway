package client

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/sony/gobreaker"
)

type CBInvoker struct {
	invoker     Invoker
	cb          *gobreaker.CircuitBreaker
	errorsCount int64
}

func NewCBInvoker(name string, invoker Invoker) *CBInvoker {
	st := gobreaker.Settings{
		Name:        name,
		MaxRequests: 5,                // сколько попыток в half-open
		Interval:    60 * time.Second, // сброс счетчиков ошибок
		Timeout:     30 * time.Second, // время, после которого breaker открывается
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3 // открываем после 3 подряд ошибок
		},
	}

	return &CBInvoker{
		invoker: invoker,
		cb:      gobreaker.NewCircuitBreaker(st),
	}
}

func (c *CBInvoker) Invoke(ctx context.Context, method string, payload []byte) ([]byte, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		resp, err := c.invoker.Invoke(ctx, method, payload)
		if err != nil {
			atomic.AddInt64(&c.errorsCount, 1)
			return nil, err
		}
		return resp, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]byte), nil
}
