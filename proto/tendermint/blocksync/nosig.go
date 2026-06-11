package blocksync

// NoSig is a zero-size stand-in for a CommitSig / ExtendedCommitSig used by
// the SigCount stub messages. Its Unmarshal accepts (and discards) any wire
// payload, and the slice-of-NoSig that gogoproto generates costs no memory
// per entry — only the slice header grows.
type NoSig struct{}

func (NoSig) Marshal() ([]byte, error)                 { return nil, nil }
func (NoSig) MarshalTo([]byte) (int, error)            { return 0, nil }
func (NoSig) MarshalToSizedBuffer([]byte) (int, error) { return 0, nil }
func (NoSig) Size() int                                { return 0 }

func (*NoSig) Unmarshal([]byte) error { return nil }
