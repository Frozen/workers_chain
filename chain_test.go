package workers_chain_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/frozen/workers_chain"
	"github.com/stretchr/testify/require"
)

type w1 struct {
}

func (w w1) Work(ch Channel) error {
	for {
		value, ok := ch.Receive()
		if !ok {
			return nil
		}
		v := value.(int)
		if v == 0 {
			return errors.New("error first worker")
		}
		ok = ch.Send(v * 10)
		if !ok {
			return nil
		}
	}
}

type w2 struct {
}

func (w w2) Work(ch Channel) error {
	for {
		value, ok := ch.Receive()
		if !ok {
			return nil
		}
		v := value.(int)
		if v == 0 {
			return errors.New("error second worker")
		}
		ch.Send(v + 2)
	}
}

func TestRunWithoutError(t *testing.T) {
	ctx := context.Background()
	q := New(ctx, 64)

	manager := q.AddWorker(w1{})

	manager.Send(1)
	manager.Interrupt()

	recv, _ := manager.Receive()
	require.Equal(t, 10, recv)

	recv, ok := manager.Receive()
	require.Equal(t, nil, recv)
	require.Equal(t, false, ok)

	err := q.Wait()
	require.NoError(t, err)
}

func TestRunWithError(t *testing.T) {
	ctx := context.Background()
	q := New(ctx, 64)

	manager := q.AddWorker(w1{})
	manager.Send(0)

	recv, ok := manager.Receive()
	require.Equal(t, nil, recv)
	require.Equal(t, false, ok)

	err := q.Wait()
	require.Error(t, err)
}

func TestSecondWorkerReceiveMessage(t *testing.T) {
	ctx := context.Background()
	q := New(ctx, 64)

	manager1 := q.AddWorker(w1{})
	manager1.Send(1)
	manager1.Interrupt()

	manager2 := q.AddWorker(w2{})

	recv1, ok1 := manager2.Receive()
	require.Equal(t, 12, recv1)
	require.Equal(t, true, ok1)

	recv2, ok2 := manager2.Receive()
	require.Equal(t, nil, recv2)
	require.Equal(t, false, ok2)

	err := q.Wait()
	require.NoError(t, err)
}

func TestSecondWorkerInterruptWithoutError(t *testing.T) {
	ctx := context.Background()
	q := New(ctx, 64)

	manager1 := q.AddWorker(w1{})
	manager1.Send(1)

	manager2 := q.AddWorker(w2{})

	recv, _ := manager2.Receive()
	require.Equal(t, 12, recv)
	manager2.Interrupt()

	require.NoError(t, q.Wait())
}

func TestSecondWorkerInterruptWithError(t *testing.T) {
	ctx := context.Background()
	q := New(ctx, 64)

	manager1 := q.AddWorker(w1{})
	manager1.Send(1)

	manager2 := q.AddWorker(w2{})

	recv, _ := manager2.Receive()
	require.Equal(t, 12, recv)
	manager2.Send(0)

	err := q.Wait()
	require.EqualError(t, err, "error second worker")
}

type producer struct {
}

func (p producer) Produce(_ context.Context, channel OutgoingChannel) error {
	for i := 1; i <= 3; i++ {
		channel.Send(i)
	}
	return nil
}

func TestWorkersWithProducer(t *testing.T) {
	ctx := context.Background()
	q := New(ctx, 64)

	q.AddWorker(w1{})
	m2 := q.AddWorker(w2{})
	q.AddProducer(producer{})

	require.Equal(t, 12, recv(m2.Receive()))
	require.Equal(t, 22, recv(m2.Receive()))
	require.Equal(t, 32, recv(m2.Receive()))

	require.NoError(t, q.Wait())
}

func recv(a interface{}, _ bool) interface{} {
	return a
}

func TestNoWorkersWait(t *testing.T) {
	ctx := context.Background()
	q := New(ctx, 64)

	require.NoError(t, q.Wait())
}

func TestRunWithCancelledContextWithoutWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	q := New(ctx, 64)
	require.NoError(t, q.Wait())
}

func TestRunWithCancelledContextWithWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	q := New(ctx, 64)
	q.AddWorker(w1{})
	require.NoError(t, q.Wait())
}
