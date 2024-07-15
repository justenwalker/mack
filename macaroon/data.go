package macaroon

import (
	"bytes"
	"unsafe"
)

const (
	macaroonDataOverhead = unsafe.Offsetof(macaroonData{}.data)
	caveatDataOverhead   = unsafe.Offsetof(caveatData{}.data)
)

type macaroonData struct {
	locSize     uint16
	idSize      uint16
	sigSize     uint16
	caveatCount uint16
	caveatSize  uint64
	data        [1]byte
}

func (m *macaroonData) size() uintptr {
	return macaroonDataOverhead + uintptr(m.locSize+m.idSize+m.sigSize) + uintptr(m.caveatSize)
}

type caveatData struct {
	idSize  uint16
	vidSize uint16
	locSize uint16
	data    [1]byte
}

func (c *caveatData) size() uintptr {
	return caveatDataOverhead + uintptr(c.vidSize+c.idSize+c.locSize)
}

func (c *caveatData) thirdParty() bool {
	return c.vidSize > 0
}

// returns the bytes for HMAC, which is concat(vId,cId).
func (c *caveatData) hmacData() []byte {
	return unsafe.Slice(&c.data[0], c.vidSize+c.idSize)
}

func (c *caveatData) vid() []byte {
	return unsafe.Slice(&c.data[0], c.vidSize)
}

func (c *caveatData) cid() []byte {
	return unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(&c.data[0]), c.vidSize)), c.idSize)
}

func (c *caveatData) loc() []byte {
	return unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(&c.data[0]), c.vidSize+c.idSize)), c.locSize)
}

func newMacaroonData(loc string, id []byte, sigSize int, rcs ...RawCaveat) *macaroonData {
	cavSize := rawCaveatsSize(rcs)
	sz := macaroonDataOverhead + uintptr(len(loc)+len(id)+sigSize) + cavSize
	data := make([]byte, sz)
	nmd := (*macaroonData)(unsafe.Pointer(&data[0]))
	nmd.locSize = uint16(len(loc))
	data = data[macaroonDataOverhead:]
	n := copy(data, loc)
	nmd.idSize = uint16(len(id))
	n += copy(data[n:n+int(nmd.idSize)], id)
	nmd.sigSize = uint16(sigSize)
	nmd.caveatCount = uint16(len(rcs))
	nmd.caveatSize = uint64(cavSize)
	cavdata := data[n:]
	for i := range rcs {
		cp := (*caveatData)(unsafe.Pointer(&cavdata[0]))
		cp.vidSize = uint16(len(rcs[i].VID))
		copy(cp.vid(), rcs[i].VID)

		cp.idSize = uint16(len(rcs[i].CID))
		copy(cp.cid(), rcs[i].CID)

		cp.locSize = uint16(len(rcs[i].Location))
		copy(cp.loc(), rcs[i].Location)

		cavdata = cavdata[cp.size():]
	}
	return nmd
}

func (m *macaroonData) clone() *macaroonData {
	bs := m.bytes()
	data := make([]byte, len(bs))
	copy(data, bs)
	return (*macaroonData)(unsafe.Pointer(&data[0]))
}

func (m *macaroonData) appendCaveats(s *Scheme, rcs ...RawCaveat) *macaroonData {
	bs := m.bytes()
	cavSize := rawCaveatsSize(rcs)
	data := make([]byte, len(bs)+int(cavSize))
	n := copy(data, bs)
	n -= int(m.sigSize)
	sig := data[len(bs)+int(cavSize)-s.keySize:]
	copy(sig, m.sig())
	cavdata := data[n:]
	for i := range rcs {
		cp := (*caveatData)(unsafe.Pointer(&cavdata[0]))
		cp.vidSize = uint16(len(rcs[i].VID))
		copy(cp.vid(), rcs[i].VID)

		cp.idSize = uint16(len(rcs[i].CID))
		copy(cp.cid(), rcs[i].CID)

		cp.locSize = uint16(len(rcs[i].Location))
		copy(cp.loc(), rcs[i].Location)

		cavdata = cavdata[cp.size():]
		if err := s.hmac.HMAC(sig, sig, cp.hmacData()); err != nil {
			panic(err) // HMAC should never fail
		}
	}
	nmd := (*macaroonData)(unsafe.Pointer(&data[0]))
	nmd.caveatCount += uint16(len(rcs))
	nmd.caveatSize += uint64(cavSize)
	return nmd
}

func (m *macaroonData) bytes() []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(m)), m.size())
}

func (m *macaroonData) loc() []byte {
	return unsafe.Slice(&m.data[0], m.locSize)
}

func (m *macaroonData) id() []byte {
	ptr := (*byte)(unsafe.Add(unsafe.Pointer(&m.data[0]), m.locSize))
	return unsafe.Slice(ptr, m.idSize)
}

func (m *macaroonData) sig() []byte {
	ptr := (*byte)(unsafe.Add(unsafe.Pointer(&m.data[0]), uintptr(m.locSize)+uintptr(m.idSize)+uintptr(m.caveatSize)))
	return unsafe.Slice(ptr, m.sigSize)
}

func (m *macaroonData) caveatStart() *byte {
	return (*byte)(unsafe.Add(unsafe.Pointer(&m.data[0]), uintptr(m.locSize+m.idSize)))
}

func (m *macaroonData) caveats() []Caveat {
	caveats := make([]Caveat, m.caveatCount)
	cavdata := m.caveatStart()
	var offset uintptr
	for i := 0; i < int(m.caveatCount); i++ {
		cp := (*caveatData)(unsafe.Add(unsafe.Pointer(cavdata), offset))
		caveats[i] = Caveat{caveatData: cp}
		offset += caveatDataOverhead + uintptr(cp.vidSize+cp.idSize+cp.locSize)
		if offset > uintptr(m.caveatSize) {
			panic("caveat size overflow")
		}
	}
	return caveats
}

func (m *macaroonData) equal(o *macaroonData) bool {
	return bytes.Equal(m.bytes(), o.bytes())
}

func rawCaveatsSize(rcs []RawCaveat) (cavSize uintptr) {
	for i := range rcs {
		cavSize += uintptr(rcs[i].size()) + caveatDataOverhead
	}
	return cavSize
}
