package dtx

import (
	"encoding/binary"
)

type FragmentDecoder struct {
	firstFragment DtxMessage
	fragments     []DtxMessage
	finished      bool
}

func NewFragmentDecoder(firstFragment DtxMessage) *FragmentDecoder {
	if !firstFragment.IsFirstFragment() {
		panic("Illegalstate, need to pass in a firstFragment")
	}
	return &FragmentDecoder{firstFragment, make([]DtxMessage, firstFragment.Fragments-1), false}
}

func (f *FragmentDecoder) AddFragment(fragment DtxMessage) bool {
	if !f.firstFragment.MessageIsFirstFragmentFor(fragment) {
		return false
	}
	f.fragments[fragment.FragmentIndex-1] = fragment
	if fragment.IsLastFragment() {
		f.finished = true
	}
	return true
}
func (f FragmentDecoder) HasFinished() bool {
	return f.finished
}

func (f FragmentDecoder) Extract() []byte {
	if !f.finished {
		panic("illegal state")
	}
	assembledMessage := make([]byte, f.firstFragment.MessageLength+32)
	copy(assembledMessage, f.firstFragment.fragmentBytes)
	//patch in correct fragment value
	binary.LittleEndian.PutUint16(assembledMessage[8:], 0)
	binary.LittleEndian.PutUint16(assembledMessage[10:], 1)
	offset := 32
	for _, frag := range f.fragments {
		copy(assembledMessage[offset:], frag.fragmentBytes)
		offset += len(frag.fragmentBytes)
	}

	//println(hex.EncodeToString(assembledMessage))
	return assembledMessage
}
