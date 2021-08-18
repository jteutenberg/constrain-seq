package align

import (
	"github.com/jteutenberg/pipe-cleaner/sequencing"
	"github.com/jteutenberg/pipe-cleaner/sequencing/rle"
)

type CostFunction interface {
	//GetCost given an input sequence and position
	GetCost(sequencing.Sequence, uint) uint
}

type fixedCost struct {
	cost uint
}

func NewFixedCost(cost uint) CostFunction {
	return &fixedCost{cost:cost}
}

func (c *fixedCost) GetCost(seq sequencing.Sequence, pos uint) uint {
	return c.cost
}

//rleCost is a simple example of applying a cost function to compressed sequences
type rleCost struct {
	costPerBase uint
}

func NewRLECost(cost uint) CostFunction {
	return &rleCost{costPerBase:cost}
}

func (c *rleCost) GetCost(seq sequencing.Sequence, pos uint) uint {
	if rleSeq, ok := seq.(rle.RLESequence); ok {
		return c.costPerBase * uint(rleSeq.GetContents()[pos])
	}
	return c.costPerBase
}
