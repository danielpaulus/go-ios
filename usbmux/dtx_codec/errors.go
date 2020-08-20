package dtx

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

func NewOutOfSync(message string) error {
	return dtxError{message, true, false}
}

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

func IsOutOfSync(err error) bool {
	te, ok := err.(outofsync)
	return ok && te.OutOfSync()
}

func IsIncomplete(err error) bool {
	te, ok := err.(incomplete)
	return ok && te.IsIncomplete()
}
