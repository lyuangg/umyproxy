package proxy

import (
	"github.com/lyuangg/umyproxy/protocol"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type (
    // for testing. implements interface protocol.Connector
    MysqlTestConn struct {}
)

func TestGet(t *testing.T) {
    p := NewPool(newOption(2))
    p.SetCreater(newTestCreater)

    conn, err := p.Get()
    assert.Nil(t, err, "pool: conn err")
    assert.NotNil(t, conn, "pool: conn is nil")

    conn2, err2 := p.Get()
    assert.Nil(t, err2, "pool: conn2 err")
    assert.NotNil(t, conn2, "pool: conn2 is nil")

    conn3, err3 := p.Get()
    assert.ErrorIs(t, err3, ErrWaitConnTimeout)
    assert.Nil(t, conn3, "pool: conn3 is not nil")
}

func TestPutConn(t *testing.T) {
    p := NewPool(newOption(1))
    p.SetCreater(newTestCreater)

    conn, err := p.Get()
    assert.Nil(t, err, "pool: conn err")
    assert.NotNil(t, conn, "pool: conn is nil")

    p.Put(conn)

    conn2, err2 := p.Get()
    assert.Nil(t, err2, "pool: conn2 err")
    assert.NotNil(t, conn2, "pool: conn2 is nil")
    assert.Equal(t, conn, conn2, "pool: conn not equal conn2")

    // wait
    go func() {
        time.Sleep(time.Millisecond * 50)
        p.Put(conn2)
    }()

    conn3, err3 := p.Get()
    assert.Nil(t, err3, "pool: conn3 err")
    assert.Equal(t, conn2, conn3, "pool: conn2 not equal conn3")
}

func newOption(num int) PoolOption {
    option := PoolOption{
        Host: "127.0.0.1",
        Port: 3306,
        PoolMaxSize: num,
        MaxLifetime: 3600 * time.Second,
        WaitTimeout: 100 * time.Millisecond,
    }
    return option
}

func newTestCreater(address string) (protocol.Connector, error) {
    c := &MysqlTestConn{}
    return c, nil
}

func (m *MysqlTestConn) ReadPacket() (protocol.Packet, error) {
    p := protocol.Packet{}
    return p, nil
}
func (m *MysqlTestConn) WritePacket(protocol.Packet) error  {
    return nil
}
func (m *MysqlTestConn) Auth(protocol.Connector) error {
    return nil
}
func (m *MysqlTestConn) TransportCmdResp(protocol.Connector) error {
    return nil
}
func (m *MysqlTestConn) Closed() bool {
    return false
}
func (m *MysqlTestConn) Expired(time.Duration) bool {
    return false
}
func (m *MysqlTestConn) RefreshUseTime() {}
func (m *MysqlTestConn) Close() error {
    return nil
}
