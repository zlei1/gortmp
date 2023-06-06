package pio

import "fmt"

type Error struct {
	N int
}

func (self Error) Error() string {
	return fmt.Sprintf("PIOFailed(%d)", self.N)
}

func ReadU8(b []byte, n *int) (v uint8, err error) {
	if len(b) < *n+1 {
		err = Error{N: *n}
		return
	}
	if b != nil {
		v = b[*n]
	}
	*n += 1
	return
}

func WriteU8(b []byte, n *int, v uint8) {
	if b != nil {
		b[*n] = v
	}
	*n += 1
	return
}

func ReadU16BE(b []byte, n *int) (v uint16, err error) {
	if len(b) < *n+2 {
		err = Error{N: *n}
		return
	}
	if b != nil {
		v = U16BE(b[*n:])
	}
	*n += 2
	return
}

func U16BE(b []byte) (i uint16) {
	i = uint16(b[0])
	i <<= 8
	i |= uint16(b[1])
	return
}

func WriteU16BE(b []byte, n *int, v uint16) {
	if b != nil {
		PutU16BE(b[*n:], v)
	}
	*n += 2
	return
}

func PutU16BE(b []byte, v uint16) {
	b[0] = byte(v >> 8)
	b[1] = byte(v)
}

func ReadU24BE(b []byte, n *int) (v uint32, err error) {
	if len(b) < *n+3 {
		err = Error{N: *n}
		return
	}
	if b != nil {
		v = U24BE(b[*n:])
	}
	*n += 3
	return
}

func U24BE(b []byte) (i uint32) {
	i = uint32(b[0])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[2])
	return
}

func WriteU24BE(b []byte, n *int, v uint32) {
	if b != nil {
		PutU24BE(b[*n:], v)
	}
	*n += 3
	return
}

func PutU24BE(b []byte, v uint32) {
	b[0] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[2] = byte(v)
}

func ReadU32BE(b []byte, n *int) (v uint32, err error) {
	if len(b) < *n+4 {
		err = Error{N: *n}
		return
	}
	if b != nil {
		v = U32BE(b[*n:])
	}
	*n += 4
	return
}

func U32BE(b []byte) (i uint32) {
	i = uint32(b[0])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[2])
	i <<= 8
	i |= uint32(b[3])
	return
}

func WriteU32BE(b []byte, n *int, v uint32) {
	if b != nil {
		PutU32BE(b[*n:], v)
	}
	*n += 4
	return
}

func PutU32BE(b []byte, v uint32) {
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}

func ReadU64BE(b []byte, n *int) (v uint64, err error) {
	if len(b) < *n+8 {
		err = Error{N: *n}
		return
	}
	if b != nil {
		v = U64BE(b[*n:])
	}
	*n += 8
	return
}

func U64BE(b []byte) (i uint64) {
	i = uint64(b[0])
	i <<= 8
	i |= uint64(b[1])
	i <<= 8
	i |= uint64(b[2])
	i <<= 8
	i |= uint64(b[3])
	i <<= 8
	i |= uint64(b[4])
	i <<= 8
	i |= uint64(b[5])
	i <<= 8
	i |= uint64(b[6])
	i <<= 8
	i |= uint64(b[7])
	return
}

func WriteU64BE(b []byte, n *int, v uint64) {
	if b != nil {
		PutU64BE(b[*n:], v)
	}
	*n += 8
	return
}

func PutU64BE(b []byte, v uint64) {
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
}

func ReadI24BE(b []byte, n *int) (v int32, err error) {
	if len(b) < *n+3 {
		err = Error{N: *n}
		return
	}
	if b != nil {
		v = I24BE(b[*n:])
	}
	*n += 3
	return
}

func I24BE(b []byte) (i int32) {
	i = int32(int8(b[0]))
	i <<= 8
	i |= int32(b[1])
	i <<= 8
	i |= int32(b[2])
	return
}

func WriteI24BE(b []byte, n *int, v int32) {
	if b != nil {
		PutI24BE(b[*n:], v)
	}
	*n += 3
	return
}

func PutI24BE(b []byte, v int32) {
	b[0] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[2] = byte(v)
}

func U32LE(b []byte) (i uint32) {
	i = uint32(b[3])
	i <<= 8
	i |= uint32(b[2])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[0])
	return
}

func WriteU32LE(b []byte, n *int, v uint32) {
	if b != nil {
		PutU32LE(b[*n:], v)
	}
	*n += 4
	return
}

func PutU32LE(b []byte, v uint32) {
	b[3] = byte(v >> 24)
	b[2] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[0] = byte(v)
}

func ReadBytes(b []byte, n *int, length int) (v []byte, err error) {
	if len(b) < *n+length {
		err = Error{N: *n}
		return
	}
	if b != nil {
		v = b[*n : *n+length]
	}
	*n += length
	return
}

func ReadString(b []byte, n *int, strlen int) (v string, err error) {
	if len(b) < *n+strlen {
		err = Error{N: *n}
		return
	}
	if b != nil {
		v = string(b[*n : *n+strlen])
	}
	*n += strlen
	return
}

func WriteBytes(b []byte, n *int, v []byte) {
	if b != nil {
		copy(b[*n:], v)
	}
	*n += len(v)
	return
}

func WriteString(b []byte, n *int, v string) {
	if b != nil {
		copy(b[*n:], v)
	}
	*n += len(v)
	return
}
