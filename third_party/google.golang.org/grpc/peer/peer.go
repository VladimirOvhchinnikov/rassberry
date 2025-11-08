package peer

import (
	"context"
	"net"
)

type Peer struct {
	Addr net.Addr
}

func FromContext(ctx context.Context) (*Peer, bool) {
	return nil, false
}
