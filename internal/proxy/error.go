package proxy

import "errors"

var (
    ErrConnExpired = errors.New("connection expired")
    ErrConnClosed = errors.New("connection Closed")
    ErrPoolClosed = errors.New("connection closed")
    ErrPoolFull = errors.New("pool full")
    ErrWaitConnTimeout = errors.New("wait mysql connection timeout")
)
