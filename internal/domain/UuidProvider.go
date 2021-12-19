package domain

import "github.com/google/uuid"

type UuidProvider struct {
}

func NewUuidProvider() UuidProvider {
	return UuidProvider{}
}

func (up UuidProvider) NewUuid() string {
	uid, _ := uuid.NewUUID()
	return uid.String()
}
