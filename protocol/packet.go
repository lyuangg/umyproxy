package protocol

const (
    MAX_PAYLOAD_LEN  int = 1<<24 - 1

	OK_PACKET  byte = 0x00
	ERR_PACKET byte = 0xff
	EOF_PACKET byte = 0xfe
    QUIT_PACKET byte= 0x01
)


type (
    Packet struct {
        // 不包含头
        Payload []byte
        SeqId uint8
    }
)

// 分包
func (p Packet) Split() []Packet {
    data := p.Payload
    pll := len(data)
    packets := make([]Packet, 0)
    seqId := p.SeqId


    for pll >= MAX_PAYLOAD_LEN {
        pk := Packet{
            Payload: data[:MAX_PAYLOAD_LEN],
            SeqId: seqId,
        }
        packets = append(packets, pk)
        data = data[MAX_PAYLOAD_LEN:]
        pll = len(data)
        seqId ++
    }

    if pll > 0 {
        pk := Packet{
            Payload: data,
            SeqId: seqId,
        }
        packets = append(packets, pk)
    } else {
        pk := Packet{}
        packets = append(packets, pk)
    }

    return packets
}

// 包头
func (p Packet) Header() []byte {
    header := make([]byte, 4)
    length := len(p.Payload)
    if length >= MAX_PAYLOAD_LEN {
        header[0] = 0xff
        header[1] = 0xff
        header[2] = 0xff
    } else {
        header[0] = byte(length)
        header[1] = byte(length >> 8)
        header[2] = byte(length >> 16)
    }
    header[3] = p.SeqId

    return header
}

func IsQuitPacket(p Packet) bool {
    if len(p.Payload) > 0 && p.Payload[0] == QUIT_PACKET {
        return true
    }
    return false
}

func IsEofPacket(p Packet) bool {
    if len(p.Payload) > 0 && p.Payload[0] == EOF_PACKET {
        return true
    }
    return false
}

func IsOkPacket(p Packet) bool {
    if len(p.Payload) > 0 && p.Payload[0] == OK_PACKET {
        return true
    }
    return false
}

func IsErrPacket(p Packet) bool {
    if len(p.Payload) > 0 && p.Payload[0] == ERR_PACKET {
        return true
    }
    return false
}
