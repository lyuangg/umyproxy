package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader(t *testing.T) {
    mockSize := []int{
        100,
        1000,
        MAX_PAYLOAD_LEN,
        MAX_PAYLOAD_LEN + 10,
    }

    for _, size := range mockSize {
        data := make([]byte, size)
        p := Packet{Payload: data, SeqId: 0}
        h := p.Header()
        assert.Len(t, h, 4, "header length error")
        if size >= MAX_PAYLOAD_LEN {
            assert.Equal(t, int(h[2]), int(0xff), "header size error")
        } else {
            assert.Equal(t, int(h[0]), int(byte(size)), "header size error")
        }
    }
}

func TestSplit(t *testing.T) {
    testCase := []struct {
        name string
        size int
        sliceLen int
    } {
        {
            "100",
            100,
            1,
        },
        {
            "1000",
            1000,
            1,
        },
        {
            "max_payload1",
            MAX_PAYLOAD_LEN,
            2,
        },
        {
            "max_payload2",
            MAX_PAYLOAD_LEN + 100,
            2,
        },
    }

    for _, tc := range testCase {
        t.Run(tc.name, func(t *testing.T) {
            data := make([]byte, tc.size)
            p := Packet{Payload: data, SeqId: 0}
            ps := p.Split()
            assert.Len(t, ps, tc.sliceLen, "split length error")
        })
    }
}
