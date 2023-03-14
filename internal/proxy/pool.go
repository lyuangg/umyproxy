package proxy

import (
	"errors"
	"fmt"
	"github.com/lyuangg/umyproxy/protocol"
	"net"
	"sync"
	"time"
)

type (
	PoolOption struct {
		Host        string
		Port        int
		MaxLifetime time.Duration
		PoolMaxSize int
		WaitTimeout time.Duration
	}

	Pool struct {
		option       PoolOption
		mu           sync.Mutex
		freeConn     []protocol.Connector
		openSize     int
		connRequests map[uint64]chan protocol.Connector
		nextRequest  uint64
		closed       bool
		createConn   ConnCreater
	}

	ConnCreater func(string) (protocol.Connector, error)
)

func NewPool(option PoolOption) *Pool {
	freeConn := make([]protocol.Connector, 0)
	connRequest := make(map[uint64]chan protocol.Connector, 0)
	var createConn ConnCreater
	createConn = NewConnect
	return &Pool{option: option, freeConn: freeConn, connRequests: connRequest, createConn: createConn}
}

func (p *Pool) SetCreater(creater ConnCreater) {
	p.createConn = creater
}

func (p *Pool) Get() (protocol.Connector, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, errors.New("pool closed")
	}

	// 空闲连接
	freeNum := len(p.freeConn)
	if freeNum > 0 {
		for i, conn := range p.freeConn {

			// 判断 conn 过期
			if !conn.Expired(p.option.MaxLifetime) && !conn.Closed() {

				// 删除
				copy(p.freeConn, p.freeConn[i+1:])
				p.freeConn = p.freeConn[:freeNum-i-1]

				conn.RefreshUseTime()

				p.mu.Unlock()
				return conn, nil
			}

			conn.Close()
			p.openSize--
		}

		// clean all
		p.freeConn = nil
	}

	// 创建新连接
	if p.openSize < p.option.PoolMaxSize {
		conn, err := p.createConn(fmt.Sprintf("%s:%d", p.option.Host, p.option.Port))
		if err != nil {
			p.mu.Unlock()
			return nil, fmt.Errorf("new connect err: %w", err)
		}
		p.openSize++
		p.mu.Unlock()
		return conn, nil
	}

	// 等待队列
	req := make(chan protocol.Connector, 1)
	reqKey := p.nextRequest + 1
	p.nextRequest = reqKey
	p.connRequests[reqKey] = req
	p.mu.Unlock()
	select {
	case <-time.After(p.option.WaitTimeout):
		p.mu.Lock()
		delete(p.connRequests, reqKey)
		p.mu.Unlock()

		// put
		select {
		default:
		case conn, ok := <-req:
			if ok && !conn.Closed() && !conn.Expired(p.option.MaxLifetime) {
				p.Put(conn)
			}
		}

		return nil, ErrWaitConnTimeout
	case conn, ok := <-req:
		if !ok {
			return nil, ErrWaitConnTimeout
		}
		return conn, nil
	}
}

func (p *Pool) Put(conn protocol.Connector) error {
	p.mu.Lock()
	if p.closed {
		conn.Close()
		p.openSize--
		p.mu.Unlock()
		return ErrPoolClosed
	}

	if conn.Expired(p.option.MaxLifetime) || conn.Closed() {
		conn.Close()
		p.openSize--
		p.mu.Unlock()
		return ErrConnExpired
	}
	conn.RefreshUseTime()

	// 请求队列
	if len(p.connRequests) > 0 {
		for reqKey, ch := range p.connRequests {
			ch <- conn
			delete(p.connRequests, reqKey)
			close(ch)
			p.mu.Unlock()
			return nil
		}
	}

	// 放入freeConn
	freeNum := len(p.freeConn)
	if freeNum >= p.option.PoolMaxSize {
		// 删掉一个
		copy(p.freeConn, p.freeConn[1:])
		p.freeConn = p.freeConn[:len(p.freeConn)-1]
	}
	p.freeConn = append(p.freeConn, conn)
	p.mu.Unlock()

	return nil
}

func (p *Pool) Close() {
	p.mu.Lock()
	p.closed = true
	for _, conn := range p.freeConn {
		p.openSize--
		conn.Close()
	}
	p.freeConn = nil
	for _, reqCh := range p.connRequests {
		close(reqCh)
	}
	p.mu.Unlock()
}

func (p *Pool) OpenSize() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.openSize
}

func NewConnect(address string) (protocol.Connector, error) {
	conn, err := net.DialTimeout("tcp", address, time.Second*2)
	if err != nil {
		return nil, fmt.Errorf("new tcp connect err: %w", err)
	}
	tcpconn := conn.(*net.TCPConn)
	tcpconn.SetKeepAlive(true)
	mysqlConn := protocol.NewConn(tcpconn)
	return mysqlConn, nil
}
