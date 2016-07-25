package websocket

import (
    "bytes"
	"encoding/binary"
	"errors"
	"io"
)

var WSErrorOutOfRange = errors.New("WS Error: Out of range, Len8(0 - 127)")

type Packet struct {
	FIN      byte
	RSV1     byte
	RSV2     byte
	RSV3     byte
	OpCode   byte
	MASK     byte
	Len8     uint8
	Len16    uint16
	Len64    uint64
	MASK_KEY [4]byte
	Buffer   bytes.Buffer
}

func EnPacket(p *Packet) (*bytes.Buffer, error) {

	var b = new(bytes.Buffer)

	var a = p.FIN | p.RSV1 | p.RSV2 | p.RSV3 | p.OpCode

	if e := b.WriteByte(a); e != nil {
        
		return nil, e
	}

	a = p.MASK | p.Len8

	if e := b.WriteByte(a); e != nil {
        
		return nil, e
	}

	switch {

	case p.Len8 == 0:

	case p.Len8 > 0 && p.Len8 <= 125:

		io.CopyN(b, &p.Buffer, int64(p.Len8))

	case p.Len8 == 126:

		if e := binary.Write(b, binary.BigEndian, p.Len16); e != nil {
            
			return nil, e
		}

		io.CopyN(b, &p.Buffer, int64(p.Len16))

	case p.Len8 == 127:

		if e := binary.Write(b, binary.BigEndian, p.Len64); e != nil {
            
			return nil, e
		}

		io.CopyN(b, &p.Buffer, int64(p.Len64))

	default:

		return nil, WSErrorOutOfRange
	}

	return b, nil
}

func DePacket(d io.Reader) (*Packet, error) {

	var t = make([]byte, 12)

	if _, e := d.Read(t[:2]); e != nil {

		return nil, e
	}

	var p = &Packet{

		FIN: t[0] & 0x80,

		RSV1: t[0] & 0x40,

		RSV2: t[0] & 0x20,

		RSV3: t[0] & 0x10,

		OpCode: t[0] & 0xF,

		MASK: t[1] & 0x80,

		Len8: t[1] & 0x7F,
	}

	switch {

	case p.Len8 == 0:

	case (p.Len8 > 0) && (p.Len8 <= 125):

		if _, e := d.Read(t[:4]); e != nil {

			return nil, e
		}

		io.CopyN(&p.Buffer, d, int64(p.Len8))

	case p.Len8 == 126:

		binary.Read(d, binary.BigEndian, &p.Len16)

		if _, e := d.Read(t[:4]); e != nil {

			return nil, e
		}

		io.CopyN(&p.Buffer, d, int64(p.Len16))

	case p.Len8 == 127:

		binary.Read(d, binary.BigEndian, &p.Len64)

		if _, e := d.Read(t[:4]); e != nil {

			return nil, e
		}

		io.CopyN(&p.Buffer, d, int64(p.Len64))

	default:

		return nil, WSErrorOutOfRange
	}

	for i, v := range t[:4] {

		p.MASK_KEY[i] = v
	}

	t = p.Buffer.Bytes()

	for i, v := range t {

		t[i] = v ^ p.MASK_KEY[i%4]
	}

	return p, nil
}