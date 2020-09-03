package dtx

import (
	"encoding/binary"
)

//FragmentDecoder collects DtxMessage fragments and merges them into a single DtxMessage when they are complete.
type FragmentDecoder struct {
	firstFragment Message
	fragments     []Message
	finished      bool
}

//NewFragmentDecoder creates a new decoder with the first fragment
func NewFragmentDecoder(firstFragment Message) *FragmentDecoder {
	if !firstFragment.IsFirstFragment() {
		panic("Illegalstate, need to pass in a firstFragment")
	}
	return &FragmentDecoder{firstFragment, make([]Message, firstFragment.Fragments-1), false}
}

//AddFragment adds fragments if they match the firstFragment this FragmentDecoder was created with.
//It returns true if the fragment was added and fals if the fragment was not matching this decoder's first fragment.
func (f *FragmentDecoder) AddFragment(fragment Message) bool {
	if !f.firstFragment.MessageIsFirstFragmentFor(fragment) {
		return false
	}
	f.fragments[fragment.FragmentIndex-1] = fragment
	if fragment.IsLastFragment() {
		f.finished = true
	}
	return true
}

//HasFinished can be used to check if all fragments have been added
func (f FragmentDecoder) HasFinished() bool {
	return f.finished
}

//Extract can be used to get an assembled DtxMessage from all the fragments. Never call this befor HasFinished is true.
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
	return assembledMessage
}
