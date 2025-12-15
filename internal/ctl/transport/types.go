package transport

import "context"

type Msg []byte

type Transport interface {
	Listen(ctx context.Context) error

	Send(ctx context.Context, msg *Msg) error
	Recv(ctx context.Context) (*Msg, error)
}
