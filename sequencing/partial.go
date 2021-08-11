package sequencing

type PartialSequence interface {
	GetName() string
	GetLength() uint

	OptionsAt(uint) []byte
	SetOptions(uint, []byte)
}

type partialSequence struct {
	name    string
	options [][]byte
}

func NewPartialSequence(name string, length uint, alphabet []byte) PartialSequence {
	ps := partialSequence{name: name, options: make([][]byte, length)}
	for i := 0; i < len(ps.options); i++ {
		ps.options[i] = alphabet //all options available by default
	}
	return &ps
}

func (ps *partialSequence) GetName() string {
	return ps.name
}
func (ps *partialSequence) GetLength() uint {
	return uint(len(ps.options))
}
func (ps *partialSequence) OptionsAt(i uint) []byte {
	if i >= uint(len(ps.options)) {
		return nil
	}
	return ps.options[i]
}
func (ps *partialSequence) SetOptions(i uint, opts []byte) {
	ps.options[i] = opts
}
