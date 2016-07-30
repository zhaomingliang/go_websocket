package websocket

import (
    "encoding/binary"
	"fmt"
	"io"
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

	if _, er = io.ReadAtLeast(rd, fr.memory, 2); er != nil {

		return Frame{}, er
	}

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

		if er = binary.Read(rd, binary.BigEndian, &fr.u16); er != nil {

			break
		}

		fallthrough

	case (fr.u7 > 0) && (fr.u7 <= 125):

		if er = binary.Read(rd, binary.BigEndian, &fr.masking); er != nil {

			break
		}

		fr.memory = make([]byte, (fr.u16 | uint16(fr.u7)>>fr.u16))

		if _, er = io.ReadFull(rd, fr.memory); er != nil {

			break
		}

		return fr, nil

	case fr.u7 == 127:

		if er = binary.Read(rd, binary.BigEndian, &fr.u64); er != nil {

			break
		}

		if er = binary.Read(rd, binary.BigEndian, &fr.masking); er != nil {

			break
		}

		const SIZE uint64 = 32 * 1024 * 1024 * 1024

		var buf = make([]byte, 0x10000)

		fr.disk = make([]*os.File, fr.u64/SIZE)

		for i, a, b, c := 0, uint64(0), SIZE, fr.u64; (a < c) && (er == nil); a += b {

			if (((c - a) / b) == 0) && ((c % b) > 0) {

				b = c % b
			}

			n := fmt.Sprintf("%s/%08x%08x", os.TempDir(), &fr.disk[i], i)

			fr.disk[i], er = os.OpenFile(n, 705, os.ModeTemporary|os.ModeSticky)

			for e, f, g := uint64(0), uint64(len(buf)), b; (e < g) && (er == nil); e += f {

				if (((g - e) / f) == 0) && ((g % f) > 0) {

					f = g % f
				}

				if _, er = io.ReadFull(rd, buf[:int(f)]); er != nil {

					break
				}

				if _, er = fr.disk[i].WriteAt(buf[:int(f)], int64(e)); er != nil {

					break
				}
			}

			i++
		}

		if er != nil {

			for i, f := range fr.disk {

				if f != nil {

					f.Close()

					os.Remove(f.Name())
				}
			}

			break
		}

		return fr, nil
	}

	return Frame{}, er
}