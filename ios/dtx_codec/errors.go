package dtx

// A set of errors for the nonblocking DtxDecoder. Should only be used in the debug proxy.
type outofsync interface {
	OutOfSync() bool
}

type incomplete interface {
	IsIncomplete() bool
}

type dtxError struct {
	errormsg   string
	outOfSync  bool
	incomplete bool
}

// NewOutOfSync should be used when the MagicBytes are wrong
func NewOutOfSync(message string) error {
	return dtxError{message, true, false}
}

// NewIncomplete when the Message was not complete
func NewIncomplete(message string) error {
	return dtxError{message, false, true}
}

func (e dtxError) Error() string {
	return e.errormsg
}

func (e dtxError) OutOfSync() bool {
	return e.outOfSync
}

func (e dtxError) IsIncomplete() bool {
	return e.incomplete
}

// IsOutOfSync returns true if err is an OutOfSync error
func IsOutOfSync(err error) bool {
	te, ok := err.(outofsync)
	return ok && te.OutOfSync()
}

// IsIncomplete returns true if the DtxMessage was incomplete
func IsIncomplete(err error) bool {
	te, ok := err.(incomplete)
	return ok && te.IsIncomplete()
}
