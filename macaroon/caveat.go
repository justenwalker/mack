package macaroon

import (
	"unsafe"
)

type Caveat struct {
	*caveatData
}

func (c *Caveat) ID() []byte {
	return c.cid()
}

func (c *Caveat) VID() []byte {
	return c.vid()
}

func (c *Caveat) Location() string {
	return string(c.loc())
}

func (c *Caveat) data() []byte {
	if c.caveatData.vidSize == 0 {
		return c.caveatData.cid()
	}
	return unsafe.Slice(&c.caveatData.data[0], c.caveatData.vidSize+c.caveatData.idSize)
}
