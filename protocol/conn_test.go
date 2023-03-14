package protocol

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type (
	// for test
	bufferConn struct {
		writeBuffer *bytes.Buffer
		readBuffer  *bytes.Buffer
	}

	strAddr struct{}
)

func (b *bufferConn) Write(p []byte) (int, error) {
	return b.writeBuffer.Write(p)
}
func (b *bufferConn) Read(p []byte) (int, error) {
	return b.readBuffer.Read(p)
}

func (b *bufferConn) Close() error {
	return nil
}

func (b *bufferConn) LocalAddr() net.Addr {
	return strAddr{}
}

func (b *bufferConn) RemoteAddr() net.Addr {
	return strAddr{}
}

func (b *bufferConn) SetDeadline(t time.Time) error      { return nil }
func (b *bufferConn) SetReadDeadline(t time.Time) error  { return nil }
func (b *bufferConn) SetWriteDeadline(t time.Time) error { return nil }

func (s strAddr) Network() string { return "" }
func (s strAddr) String() string  { return "" }

func NewBufferConn(write []byte, read []byte) *bufferConn {
	return &bufferConn{bytes.NewBuffer(write), bytes.NewBuffer(read)}
}

func TestWritePacket(t *testing.T) {
	writeByte := make([]byte, 5)
	writePacket := Packet{
		Payload: []byte{0x01},
		SeqId:   0,
	}
	bconn := NewBufferConn(writeByte, nil)

	c := NewConn(bconn)
	assert.Nil(t, c.WritePacket(writePacket), "write err")
}

func TestReadPacket(t *testing.T) {
	readByte := []byte{
		0x02, 0x00, 0x00, 0x00, 0x01, 0x02,
	}
	bconn := NewBufferConn(nil, readByte)
	c := NewConn(bconn)
	p, err := c.ReadPacket()

	assert.Nil(t, err, "read error")
	assert.Equal(t, p.Payload, []byte{0x01, 0x02}, "read content err")
}

func TestAuthClient(t *testing.T) {
	initBytes := []byte{
		78, 0, 0, 0,
		10, 53, 46, 55, 46, 50, 54, 45, 108, 111, 103, 0, 83, 7, 0, 0, 28, 124, 109, 25, 120, 114, 73, 7, 0,
		255, 247, 8, 2, 0, 255, 129, 21, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 69, 24, 56, 38, 82, 55, 1, 75, 127, 96, 68, 20, 0,
		109, 121, 115, 113, 108, 95, 110, 97, 116, 105, 118, 101, 95, 112, 97, 115, 115, 119, 111, 114, 100, 0,
	}
	authBytes := []byte{
		130, 0, 0, 1,
		141, 162, 27, 0, 0, 0, 0, 192, 33, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		114, 111, 111, 116, 0, 20, 35, 230, 165, 124, 231, 114, 146, 184, 193, 30, 201, 35, 24, 107,
		19, 192, 139, 5, 173, 7, 116, 101, 115, 116, 0, 109, 121, 115, 113, 108, 95, 110, 97, 116, 105,
		118, 101, 95, 112, 97, 115, 115, 119, 111, 114, 100, 0, 44, 12, 95, 99, 108, 105, 101, 110,
		116, 95, 110, 97, 109, 101, 7, 109, 121, 115, 113, 108, 110, 100, 12, 95, 115, 101, 114, 118,
		101, 114, 95, 104, 111, 115, 116, 9, 108, 111, 99, 97, 108, 104, 111, 115, 116,
	}
	authOkBytes := []byte{
		7, 0, 0, 2,
		0, 0, 0, 2, 0, 0, 0,
	}

	server := NewBufferConn(make([]byte, 1024), append(initBytes, authOkBytes...))
	client := NewBufferConn(make([]byte, 1024), authBytes)

	serverConn := NewConn(server)
	clientConn := NewConn(client)

	err := serverConn.Auth(clientConn)
	assert.Nil(t, err, "auth error")
}
