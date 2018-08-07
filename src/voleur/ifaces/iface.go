package ifaces

import ()

type UpdateType int

const ( // iota is reset to 0
	Update UpdateType = iota // c0 == 0
	Add
	Remove
)

type VoleurUpdateType struct {
	Name      string
	Vol       int
	IsSinkVol bool
	Type      UpdateType
	si_number string
}

type IAudioInterface interface {
	Listen(chan []byte)
	ApplyChanges(chan VoleurUpdateType)
}
