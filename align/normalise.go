package align

import (
	"errors"
	"github.com/jteutenberg/pipe-cleaner/pipeline"
	"github.com/jteutenberg/pipe-cleaner/sequencing"
	"sort"
)

type state struct {
	output   byte
	cost     uint
	position uint
	prev     *state
	ancestor *state
}

type stateList []*state

func (d stateList) Len() int { return len(d) }
func (d stateList) Less(i, j int) bool {
	if d[i].position != d[j].position {
		return d[i].position < d[j].position
	}
	return d[i].cost < d[j].cost
}
func (d stateList) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

type aligner struct {
	alphabet []byte

	constraints []Constraint
	insertCost  uint
	deleteCost  uint

	out         chan sequencing.Sequence
	input       <-chan sequencing.Sequence
	numRoutines int
}

func NewAligner(n int, alphabet []byte, cs []Constraint, insertCost, deleteCost uint) *aligner {
	return &aligner{out: make(chan sequencing.Sequence, n+1), alphabet: alphabet, constraints: cs, insertCost: insertCost + 4, deleteCost: deleteCost + 4, numRoutines: n}
}

func (r *aligner) GetOutput() <-chan sequencing.Sequence {
	return r.out
}

func (r *aligner) Attach(p pipeline.PipelineComponent) error {
	if producer, ok := p.(sequencing.SequenceComponent); ok {
		r.input = producer.GetOutput()
	} else {
		return errors.New("Aligner was not attached to a SequenceComponent")
	}
	return nil
}

func (r *aligner) Run(complete chan<- bool) {
	states := make([]*state, 300)
	nextStates := make([]*state, 300)
	for seq := range r.input {
		alignedSeq := r.Align(seq, states, nextStates)
		r.out <- alignedSeq
	}
	complete <- true
}

func (r *aligner) GetNumRoutines() int {
	return r.numRoutines
}
func (r *aligner) Align(seq sequencing.Sequence, states stateList, nextStates stateList) sequencing.Sequence {
	options := make([]bool, len(r.alphabet)) // temporary data
	// clear current states
	states = states[:1]
	states[0] = nil
	contents := seq.GetContents()

	thresholdCost := uint(len(contents)) * r.insertCost

	// search
	for i, b := range contents {
		alphaIndex := 0
		for ; alphaIndex < len(r.alphabet) && b != r.alphabet[alphaIndex]; alphaIndex++ {
		}

		// clear next states
		nextStates = nextStates[:0]
		// test each prior state
		for _, s := range states {
			prevCost := uint(0)
			pos := uint(0)
			var prevB byte
			var ancestor *state
			if s != nil {
				prevCost = s.cost
				prevB = s.output
				pos = s.position + 1
				ancestor = s.ancestor
			}
			// add the insert option: ignore this entirely: same as prior state, but higher cost
			if pos > 0 {
				n := state{output: prevB, cost: prevCost + r.insertCost, position: pos - 1, prev: s, ancestor: ancestor}
				nextStates = append(nextStates, &n)
			} else {
				// we can skip the first base too (i.e. it was inserted)
				// this kind of wants to be a nil state, but with a cost. So a hack.
				n := state{output: 0, cost: prevCost + r.insertCost, position: 0, prev: s, ancestor: ancestor}
				nextStates = append(nextStates, &n)
			}

			// is a step a valid match?
			for j := 0; j < len(options); j++ {
				options[j] = true
			}
			for _, c := range r.constraints {
				c.FilterOptions(pos, prevB, options)
			}
			if options[alphaIndex] {
				optionCost := uint(0)
				for j := 0; j < len(options); j++ {
					if options[j] {
						optionCost++
					}
				}

				//valid steps are free
				cost := prevCost + optionCost - 1
				nextStates = append(nextStates, &state{output: b, cost: prevCost + optionCost - 1, position: pos, prev: s, ancestor: ancestor})
				if cost+r.insertCost*uint(len(contents)-i-1) < thresholdCost {
					thresholdCost = cost + r.insertCost*uint(len(contents)-i-1)
				}
			}
			// a delete of 1-4 valid options, plus a match to a later state
			priorBase := byte(0)
			// quick test whether the deleted base has only one option: then use that base as prior
			for j, opt := range options {
				if opt {
					if priorBase == 0 {
						priorBase = r.alphabet[j]
					} else {
						priorBase = byte('N')
					}
				}
			}

			for j := pos + 1; j < pos+5; j++ {
				// make the delete state just before this
				s = &state{output: priorBase, position: j - 1, cost: prevCost + r.deleteCost, prev: s, ancestor: ancestor}
				prevCost = s.cost
				// test the options here for a match
				for k := 0; k < len(options); k++ {
					options[k] = true
				}
				for _, c := range r.constraints {
					c.FilterOptions(j, priorBase, options)
				}
				if options[alphaIndex] {
					optionCost := uint(0)
					for k := 0; k < len(options); k++ {
						if options[k] {
							optionCost++
						}
					}
					//this is a valid delete-to target
					n := &state{output: b, position: j, cost: s.cost + optionCost - 1, prev: s, ancestor: ancestor}
					nextStates = append(nextStates, n)
				}
				priorBase = byte('N') // TODO: we can check for single-option bases here too
			}

		}
		// if same position, higher cost, remove it (we are at same content x position, so is equivalent)
		sort.Sort(nextStates)
		// so keep the first of each position
		pPos := uint(1000000)
		states = states[:0]
		for _, s := range nextStates {
			if pPos != s.position {
				if s.cost <= thresholdCost {
					states = append(states, s)
				}
			}
			pPos = s.position
		}

	}
	// trace back to construct the final sequence
	outputSeq := r.traceBack(seq.GetName()+"_aligned", states)
	return outputSeq
}

func (r *aligner) traceBack(label string, states stateList) sequencing.Sequence {
	if len(states) == 0 {
		return sequencing.NewFastA(label, []byte{})
	}
	// find best end state
	bestState := states[0]
	for _, s := range states {
		if s.cost < bestState.cost {
			bestState = s
		}
	}
	//run backwards, building contents in reverse
	contents := make([]byte, bestState.position+1)
	nBase := byte('N')
	for i := 0; i < len(contents); i++ {
		contents[i] = nBase
	}
	for ; bestState != nil; bestState = bestState.prev {
		if bestState.output != 0 {
			contents[bestState.position] = bestState.output
		}
	}
	return sequencing.NewFastA(label, contents)
}
func (r *aligner) Close() {
	close(r.out)
}
