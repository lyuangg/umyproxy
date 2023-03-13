package protocol

import "fmt"

type (
	Responser interface {
		ResponsePacket(client Connector) error
	}

	QueryResponse struct {
		server Connector
	}

	UtilityResponse struct {
		server Connector
		cmd    byte
	}

	PreparedResponse struct {
		server Connector
		cmd    byte
	}
)

func NewResponse(server Connector, cmd byte) Responser {
	switch cmd {
	case COM_QUERY:
		return &QueryResponse{server: server}
	case COM_STMT_PREPARE, COM_STMT_EXECUTE, COM_STMT_CLOSE, COM_STMT_RESET, COM_STMT_SEND_LONG_DATA:
		return &PreparedResponse{server: server, cmd: cmd}
	default:
		return &UtilityResponse{server: server, cmd: cmd}
	}
}

func TransportPacket(src, dst Connector) (Packet, error) {
	p, err := src.ReadPacket()
	if err != nil {
		return p, fmt.Errorf("read src err: %w", err)
	}

	err = dst.WritePacket(p)
	if err != nil {
		return p, fmt.Errorf("write dst err: %w", err)
	}

	return p, nil
}

func (r *QueryResponse) ResponsePacket(client Connector) error {
	columnEnd := false
	for {
		p, err := TransportPacket(r.server, client)
		if err != nil {
			return err
		}

		if IsOkPacket(p) || IsErrPacket(p) {
			return nil
		}

		if IsEofPacket(p) {
			if columnEnd {
				// data end
				return nil
			} else {
				columnEnd = true
			}
		}
	}
}

func (r *UtilityResponse) ResponsePacket(client Connector) error {
	if r.cmd == COM_QUIT {
		return ErrClientQuit
	}

	if r.cmd == COM_FIELD_LIST {
		for {
			p, err := TransportPacket(r.server, client)
			if err != nil {
				return err
			}
			if IsEofPacket(p) || IsErrPacket(p) {
				return nil
			}
		}
	}

	if r.cmd == COM_STATISTICS {
		_, err := TransportPacket(r.server, client)
		return err
	}

	if r.cmd == COM_CHANGE_USER {
		_, err := TransportPacket(r.server, client)
		return err
	}

	queryResp := &QueryResponse{server: r.server}
	return queryResp.ResponsePacket(client)
}

func (r *PreparedResponse) ResponsePacket(client Connector) error {
	if r.cmd == COM_STMT_CLOSE || r.cmd == COM_STMT_SEND_LONG_DATA {
		return nil
	}

	// 预处理sql响应
	if r.cmd == COM_STMT_PREPARE {
		p, err := TransportPacket(r.server, client)
		if err != nil {
			return err
		}
		// 响应成功
		if IsOkPacket(p) {
			// columns， parameters 都为 0
			if len(p.Payload) > 8 && p.Payload[5] == 0 && p.Payload[6] == 0 && p.Payload[7] == 0 && p.Payload[8] == 0 {
				return nil
			}
			eofCount := 0

			// columns 为0
			if len(p.Payload) > 8 && p.Payload[5] == 0 && p.Payload[6] == 0 {
				eofCount = 1
			}
			// parameters 为0
			if len(p.Payload) > 8 && p.Payload[7] == 0 && p.Payload[8] == 0 {
				eofCount = 1
			}
			for {
				p2, err2 := TransportPacket(r.server, client)
				if err2 != nil {
					return err2
				}
				if IsEofPacket(p2) {
					if eofCount == 1 {
						return nil
					} else {
						eofCount++
					}
				}
			}
		}
		return nil
	}

	// 预处理语句响应
	if r.cmd == COM_STMT_EXECUTE {
		p, err := TransportPacket(r.server, client)
		if err != nil {
			return err
		}
		if IsErrPacket(p) || IsOkPacket(p) {
			return nil
		}
		eofCount := 0
		for {
			p2, err2 := TransportPacket(r.server, client)
			if err2 != nil {
				return err2
			}
			if IsEofPacket(p2) {
				if eofCount == 1 {
					return nil
				} else {
					eofCount++
				}
			}
		}
	}

	// 响应失败
	_, err := TransportPacket(r.server, client)
	return err
}
