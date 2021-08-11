package align

import (
	"github.com/jteutenberg/constrain-seq/sequencing"
)

type Constraint interface {
	// Filter valid options as a set over the alphabet, give target sequence position and prior letter
	FilterOptions(uint, byte, []bool)
}

type partialSequenceConstraint struct {
	ps       sequencing.PartialSequence
	alphabet []byte
	//TODO: create filters at each position, i.e. [][]bool
}

type noHomopolymerConstraint struct {
	alphabet []byte
}

type maxLengthConstraint struct {
	maxLength uint
}

func NewNoHomopolymerConstraint(alphabet []byte) Constraint {
	return &noHomopolymerConstraint{alphabet: alphabet}
}

func (c *noHomopolymerConstraint) FilterOptions(i uint, prev byte, options []bool) {
	for i, a := range c.alphabet {
		options[i] = options[i] && a != prev
	}
}

func NewPartialSequenceConstraint(seq sequencing.PartialSequence, alphabet []byte) Constraint {
	return &partialSequenceConstraint{ps: seq, alphabet: alphabet}
}

func (c *partialSequenceConstraint) FilterOptions(i uint, prev byte, options []bool) {
	opts := c.ps.OptionsAt(i)
	if opts == nil {
		for i := 0; i < len(options); i++ {
			options[i] = false
		}
		return
	}
	if len(opts) == len(c.alphabet) {
		return
	}
	for i, a := range c.alphabet {
		valid := false
		for _, b := range opts {
			valid = valid || b == a
		}
		options[i] = options[i] && valid
	}
}

func NewMaxLengthConstraint(maxLength uint) Constraint {
	return &maxLengthConstraint{maxLength: maxLength}
}

func (c *maxLengthConstraint) FilterOptions(i uint, prev byte, options []bool) {
	if i > c.maxLength {
		for i, _ := range options {
			options[i] = false
		}
	}
}
