package types

import "strconv"

// TokenCreationDate The field of an external purchase token that contains the UNIX date, in milliseconds, when the system created the token.
type TokenCreationDate int64

func (t TokenCreationDate) String() string {
	return strconv.FormatInt(int64(t), 10)
}

func (t TokenCreationDate) Int64() int64 {
	return int64(t)
}
