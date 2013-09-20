package grunway

import (
	"io"
	"net/http"
	"time"
)

// automatically implemented when using grunway.Entity as an anonymous field
type EntityInterface interface {
	EntitySerializer
	EntityPrimitives
}

type EntitySerializer interface {
	Serialize(writer io.Writer) error
	Deserialize(req *http.Request, route Route) error
}

type EntityPrimitives interface {
	GetPrimaryKey() int64
	SetPrimaryKey(pkey int64)
	GetVersion() int32
	SetVersion(int32)
	GetUpdatedTime() time.Time
	SetUpdatedTime(modTime time.Time)
	GetCreatedTime() time.Time
	SetCreatedTime(modTime time.Time)
}

type Entity struct {
	PrimaryKey int64     `json:",omitempty"`
	Version    uint16    `json:",omitempty"`
	Updated    time.Time `json:",omitempty"`
	Created    time.Time `json:",omitempty"`
}

func (e *Entity) SetPrimaryKey(pkey int64) {
	e.PrimaryKey = pkey
}
func (e *Entity) GetPrimaryKey() int64 {
	return e.PrimaryKey
}
func (e *Entity) GetVersion() uint16 {
	return e.Version
}
func (e *Entity) SetVersion(vTag uint16) {
	e.Version = vTag
}
func (e *Entity) GetUpdatedTime() time.Time {
	return e.Updated
}
func (e *Entity) SetUpdatedTime(newValue time.Time) {
	e.Updated = newValue
}
func (e *Entity) GetCreatedTime() time.Time {
	return e.Created
}
func (e *Entity) SetCreatedTime(newValue time.Time) {
	e.Created = newValue
}
