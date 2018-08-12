package ifaces

import ()

type UpdateType int

const ( // iota is reset to 0
	AddOrUpdate UpdateType = iota // c0 == 0
	Remove
)

type VoleurUpdateType struct {
	Name    string				`json:"name"`
	Val     int					`json:"val"`
	Type    UpdateType			`json:"type"`
	UID		string				`json:"uid"`
	AuxData map[string]string	`json:"auxdata"`
}

type IControlInterface interface {
	Listen(chan []byte)
	ApplyChanges(chan VoleurUpdateType)
	GetAll() [][]byte
}
