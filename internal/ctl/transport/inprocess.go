package transport

import (
	"context"
)

type InProcessTransport struct {
	connectedPeer *InProcessTransport

	cListening chan struct{}
	cRecv      chan *Msg
}

var _ Transport = (*InProcessTransport)(nil)

func NewInProcessTransport() *InProcessTransport {
	return &InProcessTransport{
		connectedPeer: nil,
		cListening:    make(chan struct{}),
		cRecv:         make(chan *Msg),
	}
}

func (t *InProcessTransport) SetPeer(peer *InProcessTransport) {
	t.connectedPeer = peer
}

func (t *InProcessTransport) Listen(ctx context.Context) error {
	if t.connectedPeer == nil {
		panic("peer not set")
	}

	close(t.cListening)

	<-ctx.Done()
	return nil
}

func (t *InProcessTransport) Send(ctx context.Context, msg *Msg) error {
	if t.connectedPeer == nil {
		panic("peer not set")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.cListening:
	}

	select {
	case t.connectedPeer.cRecv <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *InProcessTransport) Recv(ctx context.Context) (*Msg, error) {
	if t.connectedPeer == nil {
		panic("peer not set")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-t.cListening:
	}

	select {
	case msg := <-t.cRecv:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
