package types

import "github.com/google/uuid"

type UUID string

func (u UUID) IsValidUUID() bool {
	_, err := uuid.Parse(string(u))
	return err == nil
}
