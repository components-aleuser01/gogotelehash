package e3x

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/telehash/gogotelehash/e3x/cipherset"
	"github.com/telehash/gogotelehash/lob"
	"github.com/telehash/gogotelehash/transports"
)

func registerEventLoggers(e *Endpoint, t *testing.T) {
	observers := ObserversFromEndpoint(e)
	observers.Register(func(e *ExchangeOpenedEvent) { t.Logf("EVENT: %s", e.String()) })
	observers.Register(func(e *ExchangeClosedEvent) { t.Logf("EVENT: %s", e.String()) })
	observers.Register(func(e *ChannelOpenedEvent) { t.Logf("EVENT: %s", e.String()) })
	observers.Register(func(e *ChannelClosedEvent) { t.Logf("EVENT: %s", e.String()) })
}

type MockExchange struct {
	mock.Mock
}

func (m *MockExchange) deliver_packet(pkt *lob.Packet) error {
	args := m.Called(pkt)
	return args.Error(0)
}

func (m *MockExchange) unregister_channel(channelId uint32) {
	m.Called(channelId)
}

func (m *MockExchange) RemoteIdent() *Ident {
	args := m.Called()
	return args.Get(0).(*Ident)
}

type pipeTransport struct {
	laddr transports.Addr
	raddr transports.Addr
	r     io.ReadCloser
	w     io.WriteCloser
}

func openPipeTransport(lid, rid string) (l, r transports.Transport) {
	r1, w1, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	r2, w2, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	l = &pipeTransport{mockAddr(lid), mockAddr(rid), r1, w2}
	r = &pipeTransport{mockAddr(rid), mockAddr(lid), r2, w1}
	return
}

func (t *pipeTransport) LocalAddresses() []transports.Addr {
	return []transports.Addr{t.laddr}
}

func (t *pipeTransport) ReadMessage(p []byte) (n int, src transports.Addr, err error) {
	var (
		lbuf [2]byte
		l    uint16
	)

	_, err = io.ReadFull(t.r, lbuf[:])
	if err != nil {
		return 0, nil, err
	}

	l = binary.BigEndian.Uint16(lbuf[:])
	p = p[:int(l)]

	_, err = io.ReadFull(t.r, p)
	if err != nil {
		return 0, nil, err
	}

	return int(l), t.raddr, nil
}

func (t *pipeTransport) WriteMessage(p []byte, dst transports.Addr) error {
	var (
		lbuf [2]byte
		lp   = lbuf[:]
		l    = uint16(len(p))
	)

	binary.BigEndian.PutUint16(lbuf[:], l)

	for len(lp) > 0 {
		n, err := t.w.Write(lp[:])
		if err != nil {
			return err
		}
		lp = lp[n:]
	}

	for len(p) > 0 {
		n, err := t.w.Write(p[:])
		if err != nil {
			return err
		}
		p = p[n:]
	}

	return nil

}

func (t *pipeTransport) Close() error {
	t.r.Close()
	t.w.Close()
	return nil
}

func makeIdent(name string) *Ident {
	key, err := cipherset.GenerateKey(0x3a)
	if err != nil {
		panic(err)
	}

	ident, err := NewIdent(cipherset.Keys{0x3a: key}, nil, []transports.Addr{
		mockAddr(name),
	})
	if err != nil {
		panic(err)
	}

	return ident
}

type mockAddr string

func (m mockAddr) Network() string {
	return "mock"
}

func (m mockAddr) String() string {
	data, err := m.MarshalJSON()
	if err != nil {
		panic(err)
	}

	return string(data)
}

func (m mockAddr) MarshalJSON() ([]byte, error) {
	var desc = struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}{
		Type: "mock",
		Name: string(m),
	}

	return json.Marshal(&desc)
}

func (a mockAddr) Equal(x transports.Addr) bool {
	b, ok := x.(mockAddr)
	if !ok {
		return false
	}
	return a == b
}
