package types

type Vote struct {
	Creator   string
	OracleID  string
	Timestamp uint64
	Data      string
	Signature string
}
