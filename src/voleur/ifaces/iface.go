package ifaces

import ()

type UpdateType int

const ( // iota is reset to 0
	AddOrUpdate UpdateType = iota // c0 == 0
	Remove
)

type VoleurUpdateType struct {
	Name    string
	Vol     int
	Type    UpdateType
	AuxData map[string]string
}

type IAudioInterface interface {
	Listen(chan []byte)
	ApplyChanges(chan VoleurUpdateType)
	GetAll() [][]byte
}
