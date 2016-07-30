package websocket

import (
    "encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
)

type Frame struct {
	fin     byte
	rsv1    byte
	rsv2    byte
	rsv3    byte
	opcode  byte
	masked  byte
	u7      uint8
	u16     uint16
	u64     uint64
	masking [4]byte
	memory  []byte
	disk    []*os.File
}

func receive(rd io.Reader) (fr Frame, er error) {

	fr.memory = make([]byte, 0x14)

	io.ReadAtLeast(rd, fr.memory, 2)

	fr.fin = 0x80 & fr.memory[0]

	fr.rsv1 = 0x40 & fr.memory[0]

	fr.rsv2 = 0x20 & fr.memory[0]

	fr.rsv3 = 0x10 & fr.memory[0]

	fr.opcode = 0x0F & fr.memory[0]

	fr.masked = 0x80 & fr.memory[1]

	fr.u7 = 0x7F & fr.memory[1]

	switch {

	case fr.u7 == 0:

	case fr.u7 == 126:

		binary.Read(rd, binary.BigEndian, &fr.u16)

		fallthrough

	case (fr.u7 > 0) && (fr.u7 <= 125):

		binary.Read(rd, binary.BigEndian, &fr.masking)

		fr.memory = make([]byte, (fr.u16 | uint16(fr.u7)>>fr.u16))

		io.ReadFull(rd, fr.memory)

	case fr.u7 == 127:

		binary.Read(rd, binary.BigEndian, &fr.u64)

		binary.Read(rd, binary.BigEndian, &fr.masking)

		const SIZE = 32 * 1024 * 1024 * 1024

		var buf = make([]byte, 0x10000)

		fr.disk = make([]*os.File, fr.u64/SIZE)

		for i, a, b, c := 0, 0, SIZE, fr.u64; (a < c) && (err == nil); a += b {

			if (((c - a) / b) == 0) && ((c % b) > 0) {

				b = c % b
			}

			n := fmt.Sprintf("%08x:%d", &fr.disk[i], i)

			fr.disk[i] = os.OpenFile(n, 705, os.ModeTemporary|os.ModeSticky)

			for e, f, g := 0, len(buf), b; (e < g) && (err == nil); e += f {

				if (((g - e) / f) == 0) && ((g % f) > 0) {

					f = g % f
				}

				io.ReadFull(rd, buf[:f])

				fr.disk[i].WriteAt(buf[:f], e)
			}

			i++
		}

		if er != nil {

			for i, v := range fr.disk {

				if v != nil {

					v.Close()
				}
			}

			fr = Frame{}
		}

	}

	return fr, er
}