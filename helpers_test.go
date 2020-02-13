package workers_chain

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFuncWorker(t *testing.T) {
	err := errors.New("some err")
	var worker Worker = FuncWorker(func(channel Channel) error {
		return err
	})

	a := New(context.Background(), 1)
	a.AddWorker(worker)
	require.Equal(t, err, a.Wait())
}

func TestFuncProducer(t *testing.T) {
	err := errors.New("some err")
	var producer Producer = FuncProducer(func(ctx context.Context, ch OutgoingChannel) error {
		return err
	})

	a := New(context.Background(), 1)
	a.AddProducer(producer)
	require.Equal(t, err, a.Wait())
}
