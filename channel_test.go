package workers_chain

import (
	"testing"
)

func TestChannelMultipleClose(t *testing.T) {
	ch := newChannel(make(chan interface{}, 10))
	ch.Close()
	ch.Close()
}
