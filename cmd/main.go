package main

import (
	"bytes"
	"context"
	"fmt"
	"gortmp/av"
	"gortmp/rtmp"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	log "github.com/sirupsen/logrus"
)

var VERSION = "1.0"

// 流
type stream struct {
	n   int64
	sub sync.Map
	pub unsafe.Pointer // 通用型的指针
}

func (s *stream) notifySub() {
	s.sub.Range(func(key, value interface{}) bool {
		ss := key.(*streamSub)
		select {
		case ss.notify <- struct{}{}:
		default:
		}
		return true
	})
}

type streamSub struct {
	notify chan struct{}
}

type streamPub struct {
	cancel func()
	gc     *gopCache
}

type gopCache struct {
	pkts  []av.Packet
	idx   int
	curst unsafe.Pointer
}

type gopCacheSnapshot struct {
	pkts []av.Packet
	idx  int
}

func (gc *gopCache) put(pkt av.Packet) {
	if pkt.IsKeyFrame {
		gc.pkts = []av.Packet{}
	}
	gc.pkts = append(gc.pkts, pkt)
	gc.idx++
	st := &gopCacheSnapshot{
		pkts: gc.pkts,
		idx:  gc.idx,
	}
	atomic.StorePointer(&gc.curst, unsafe.Pointer(st))
}

func (gc *gopCache) curSnapshot() *gopCacheSnapshot {
	return (*gopCacheSnapshot)(atomic.LoadPointer(&gc.curst))
}

type mergeSeqhdr struct {
	cb     func(av.Packet)
	hdrpkt av.Packet
}

func (m *mergeSeqhdr) do(pkt av.Packet) {
	switch pkt.Type {
	case av.H264DecoderConfig:
		m.hdrpkt.VSeqHdr = append([]byte(nil), pkt.Data...)
	case av.H264:
		pkt.Metadata = m.hdrpkt.Metadata
		if pkt.IsKeyFrame {
			pkt.VSeqHdr = m.hdrpkt.VSeqHdr
		}
		m.cb(pkt)
	case av.AACDecoderConfig:
		m.hdrpkt.ASeqHdr = append([]byte(nil), pkt.Data...)
	case av.AAC:
		pkt.Metadata = m.hdrpkt.Metadata
		pkt.ASeqHdr = m.hdrpkt.ASeqHdr
		m.cb(pkt)
	case av.Metadata:
		m.hdrpkt.Metadata = pkt.Data
	}
}

func (s *stream) setPub(r av.PacketReader) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sp := &streamPub{
		cancel: cancel,
		gc:     &gopCache{},
	}

	oldsp := (*streamPub)(atomic.SwapPointer(&s.pub, unsafe.Pointer(sp)))
	if oldsp != nil {
		oldsp.cancel()
	}

	seqmerge := mergeSeqhdr{
		cb: func(pkt av.Packet) {
			sp.gc.put(pkt)
			s.notifySub()
		},
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		pkt, err := r.ReadPacket()
		if err != nil {
			return
		}

		seqmerge.do(pkt)
	}
}

type gopCacheReadCursor struct {
	lastidx int
}

func (rc *gopCacheReadCursor) advance(cur *gopCacheSnapshot) []av.Packet {
	lastidx := rc.lastidx
	rc.lastidx = cur.idx
	if diff := cur.idx - lastidx; diff <= len(cur.pkts) {
		return cur.pkts[len(cur.pkts)-diff:]
	} else {
		return cur.pkts
	}
}

type splitSeqhdr struct {
	cb     func(av.Packet) error
	hdrpkt av.Packet
}

func (s *splitSeqhdr) sendmeta(pkt av.Packet) error {
	if bytes.Compare(s.hdrpkt.Metadata, pkt.Metadata) != 0 {
		if err := s.cb(av.Packet{
			Type: av.Metadata,
			Data: pkt.Metadata,
		}); err != nil {
			return err
		}
		s.hdrpkt.Metadata = pkt.Metadata
	}
	return nil
}

func (s *splitSeqhdr) do(pkt av.Packet) error {
	switch pkt.Type {
	case av.H264:
		if err := s.sendmeta(pkt); err != nil {
			return err
		}
		if pkt.IsKeyFrame {
			if bytes.Compare(s.hdrpkt.VSeqHdr, pkt.VSeqHdr) != 0 {
				if err := s.cb(av.Packet{
					Type: av.H264DecoderConfig,
					Data: pkt.VSeqHdr,
				}); err != nil {
					return err
				}
				s.hdrpkt.VSeqHdr = pkt.VSeqHdr
			}
		}
		return s.cb(pkt)
	case av.AAC:
		if err := s.sendmeta(pkt); err != nil {
			return err
		}
		if bytes.Compare(s.hdrpkt.ASeqHdr, pkt.ASeqHdr) != 0 {
			if err := s.cb(av.Packet{
				Type: av.AACDecoderConfig,
				Data: pkt.ASeqHdr,
			}); err != nil {
				return err
			}
			s.hdrpkt.ASeqHdr = pkt.ASeqHdr
		}
		return s.cb(pkt)
	}
	return nil
}

func (s *stream) addSub(close <-chan bool, w av.PacketWriter) {
	ss := &streamSub{
		notify: make(chan struct{}, 1),
	}

	s.sub.Store(ss, nil)
	defer s.sub.Delete(ss)

	var cursor *gopCacheReadCursor
	var lastsp *streamPub

	seqsplit := splitSeqhdr{
		cb: func(pkt av.Packet) error {
			return w.WritePacket(pkt)
		},
	}

	for {
		var pkts []av.Packet

		sp := (*streamPub)(atomic.LoadPointer(&s.pub))
		if sp != lastsp {
			cursor = &gopCacheReadCursor{}
			lastsp = sp
		}
		if sp != nil {
			cur := sp.gc.curSnapshot()
			if cur != nil {
				pkts = cursor.advance(cur)
			}
		}

		if len(pkts) == 0 {
			select {
			case <-close:
				return
			case <-ss.notify:
			}
		} else {
			for _, pkt := range pkts {
				if err := seqsplit.do(pkt); err != nil {
					return
				}
			}
		}
	}
}

// 流的集合
type streams struct {
	l sync.RWMutex // 流的数量
	m map[string]*stream
}

func newStreams() *streams {
	return &streams{
		m: map[string]*stream{},
	}
}

// 添加流
func (ss *streams) add(k string) (*stream, func()) {
	ss.l.Lock()
	defer ss.l.Unlock()

	log.Println("stream", k, "add")

	s, ok := ss.m[k]
	if !ok {
		s = &stream{}
		ss.m[k] = s
	}
	s.n++

	// 返回删除流的操作方法
	return s, func() {
		log.Println("stream", k, "remove")

		ss.l.Lock()
		defer ss.l.Unlock()

		s.n--
		if s.n == 0 {
			delete(ss.m, k)
		}
	}
}

func startRtmp() error {
	rtmpAddr := ":1935"

	listen, err := net.Listen("tcp", rtmpAddr)
	if err != nil {
		return err
	}

	s := rtmp.NewServer()
	s.LogEvent = func(c *rtmp.Conn, nc net.Conn, e int) {
		eventStr := rtmp.EventString[e]
		fmt.Println("RtmpEvent", nc.LocalAddr(), nc.RemoteAddr(), eventStr)
	}

	streams := newStreams()

	s.HandleConn = func(c *rtmp.Conn, nc net.Conn) {
		stream, removeFn := streams.add(c.URL.Path)
		defer removeFn()

		if c.Publishing {
			stream.setPub(c)
		} else {
			stream.addSub(c.CloseNotify(), c)
		}
	}

	for {
		netConn, err := listen.Accept()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		go s.HandleNetConn(netConn)
	}

	return nil
}

func main() {
	log.Infof(`
 __   __   __  _|_   __ __   __
(__| (__) |  '  |_, |  )  ) |__)
 __/                        |
        version: %s
	`, VERSION)

	startRtmp()
}
