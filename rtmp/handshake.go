package rtmp

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"gortmp/util/pio"
	"io"
)

var (
	hsClientFullKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'P', 'l', 'a', 'y', 'e', 'r', ' ',
		'0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	hsServerFullKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'M', 'e', 'd', 'i', 'a', ' ',
		'S', 'e', 'r', 'v', 'e', 'r', ' ',
		'0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	hsClientPartialKey = hsClientFullKey[:30]
	hsServerPartialKey = hsServerFullKey[:36]
)

func hsMakeDigest(key []byte, src []byte, gap int) (dst []byte) {
	h := hmac.New(sha256.New, key)
	if gap <= 0 {
		h.Write(src)
	} else {
		h.Write(src[:gap])
		h.Write(src[gap+32:])
	}
	return h.Sum(nil)
}

func hsCalcDigestPos(p []byte, base int) (pos int) {
	for i := 0; i < 4; i++ {
		pos += int(p[base+i])
	}
	pos = (pos % 728) + base + 4
	return
}

func hsFindDigest(p []byte, key []byte, base int) int {
	gap := hsCalcDigestPos(p, base)
	digest := hsMakeDigest(key, p, gap)
	if bytes.Compare(p[gap:gap+32], digest) != 0 {
		return -1
	}
	return gap
}

func hsParse1(p []byte, peerkey []byte, key []byte) (ok bool, digest []byte) {
	var pos int
	if pos = hsFindDigest(p, peerkey, 772); pos == -1 {
		if pos = hsFindDigest(p, peerkey, 8); pos == -1 {
			return
		}
	}
	ok = true
	digest = hsMakeDigest(key, p[pos:pos+32], -1)
	return
}

func hsCreate01(p []byte, time uint32, ver uint32, key []byte) {
	p[0] = 3
	p1 := p[1:]
	rand.Read(p1[8:])
	pio.PutU32BE(p1[0:4], time)
	pio.PutU32BE(p1[4:8], ver)
	gap := hsCalcDigestPos(p1, 8)
	digest := hsMakeDigest(key, p1, gap)
	copy(p1[gap:], digest)
}

func hsCreate2(p []byte, key []byte) {
	rand.Read(p)
	gap := len(p) - 32
	digest := hsMakeDigest(key, p, gap)
	copy(p[gap:], digest)
}

// RTMP 协议规范并没有限定死 C0，C1，C2 和 S0，S1，S2 的顺序，但是制定了以下规则
// 客户端必须收到服务端发来的 S1 后才能发送 C2
// 客户端必须收到服务端发来的 S2 后才能发送其他数据
// 服务端必须收到客户端发来的 C0 后才能发送 S0 和 S1
// 服务端必须收到客户端发来的 C1 后才能发送 S2
// 服务端必须收到客户端发来的 C2 后才能发送其他数据
//
// 客户端和服务端握手过程
// 客户端向服务端发送 C0 和 C1 消息
// 服务端向客户端发送 S0，S1 和 S2 消息
// 客户端向服务端发送 C2 消息
//
// C0 和 S0：1个字节长度，该消息指定了 RTMP 版本号。取值范围 0~255，我们只需要知道 3 才是我们需要的就行
// C1 和 S1：1536个字节长度，由 时间戳+零值+随机数据 组成，握手过程的中间包
// C2 和 S2：1536个字节长度，由 时间戳+时间戳2+随机数据回传 组成，基本上是 C1 和 S1 的 echo 数据。一般在实现上，会令 S2 = C1，C2 = S1
// 时间戳 4 bytes 本字段包含一个发送时间戳（取值可以为零或其他任意值）
// 零值 4 bytes 本字段必须全为零
// 随机数据 1528 bytes (key + digest)
//
//	key 结构
//	  random-data: (offset) bytes
//	  key-data: 128 bytes
//	  random-data: (764 - offset - 128 - 4) bytes
//	  offset: 4 bytes
//	digest 结构
//	  offset: 4 bytes
//	  random-data: (offset) bytes
//	  digest-data: 32 bytes
//	  random-data: (764 - 4 - offset - 32) bytes
func (c *Conn) handshakeServer() (err error) {
	var random [(1 + 1536*2) * 2]byte

	// C0: 1 bytes
	// C1 1536 bytes
	// C2 1535 bytes
	C0C1C2 := random[:1536*2+1]
	C0 := C0C1C2[:1]
	C1 := C0C1C2[1 : 1536+1]
	C2 := C0C1C2[1536+1:]

	// S0: 1 bytes
	// S1: 1536 bytes
	// S2: 1536 bytes
	S0S1S2 := random[1536*2+1:]
	S0 := S0S1S2[:1]
	S1 := S0S1S2[1 : 1536+1]
	S0S1 := S0S1S2[:1536+1]
	S2 := S0S1S2[1536+1:]

	// < C0
	if _, err = io.ReadFull(c.wrapRW.rw, C0); err != nil {
		return
	}

	if C0[0] != 3 {
		err = fmt.Errorf("VersionInvalid(%d)", C0[0])
		return
	}

	// < C1
	if _, err = io.ReadFull(c.wrapRW.rw, C1); err != nil {
		return
	}

	S0[0] = 3

	clientTime := pio.U32BE(C1[0:4])
	serverTime := clientTime
	serverVersion := uint32(0x0d0e0a0d)

	var ok bool
	var digest []byte
	if ok, digest = hsParse1(C1, hsClientPartialKey, hsServerFullKey); ok {
		hsCreate01(S0S1, serverTime, serverVersion, hsServerPartialKey)
		hsCreate2(S2, digest)
	} else {
		copy(S1, C2)
		copy(S2, C1)
	}

	// > S0S1S2
	if _, err = c.wrapRW.rw.Write(S0S1S2); err != nil {
		return
	}
	if err = c.flushWrite(); err != nil {
		return
	}

	// < C2
	if _, err = io.ReadFull(c.wrapRW.rw, C2); err != nil {
		return
	}

	return
}
