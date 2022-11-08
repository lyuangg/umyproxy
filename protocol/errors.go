package protocol

import "errors"

var (
    ErrConnClosed = errors.New("connection is closed")
    ErrNoAuth = errors.New("client no auth")
    ErrAuth = errors.New("client auth error")
    ErrClientQuit = errors.New("client quit cmd")
)
