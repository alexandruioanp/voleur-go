package ifaces

import ()

type UpdateType int

const ( // iota is reset to 0
	AddOrUpdate UpdateType = iota // c0 == 0
	Remove
)

type VoleurUpdateType struct {
	Name    string				`json:"name"`
	Vol     int					`json:"vol"`
	Type    UpdateType			`json:"type"`
	AuxData map[string]string	`json:"auxdata"`
}

type IAudioInterface interface {
	Listen(chan []byte)
	ApplyChanges(chan VoleurUpdateType)
	GetAll() [][]byte
}
