package main

import (
	"flag"
	"github.com/jteutenberg/constrain-seq/align"
	"github.com/jteutenberg/constrain-seq/sequencing"
	"github.com/jteutenberg/pipe-cleaner/pipeline"
	seq "github.com/jteutenberg/pipe-cleaner/sequencing"
	"github.com/jteutenberg/pipe-cleaner/util"
	"io/ioutil"
)

func loadPartial(filename string, alphabet []byte) (sequencing.PartialSequence, error) {
	if seq, err := ioutil.ReadFile(filename); err == nil {
		partialSeq := sequencing.NewPartialSequence("template", uint(len(seq)), alphabet)
		for i, b := range seq {
			if b != 'N' {
				partialSeq.SetOptions(uint(i), []byte{b}) // single fixed option
			}
		}
		return partialSeq, nil
	} else {
		return nil, err
	}
	return nil, nil
}

func main() {
	var inputFile = flag.String("i", "", "input fasta file name")
	var partialFile = flag.String("p", "", "partial sequence text file name")
	var maxLength = flag.Uint("m", 100, "maximum sequence length")
	var threads = flag.Int("t", 1, "number of threads to use")
	var outputFile = flag.String("o", "", "output fasta file name")
	flag.Parse()

	//prepare the aligner
	alphabet := []byte{'A', 'C', 'G', 'T'}
	var constraint align.Constraint
	if targetSequence, err := loadPartial(*partialFile, alphabet); err == nil {
		constraint = align.NewPartialSequenceConstraint(targetSequence, alphabet)
	}
	aligner := align.NewAligner(*threads, alphabet, []align.Constraint{constraint, align.NewNoHomopolymerConstraint(alphabet), align.NewMaxLengthConstraint(*maxLength)}, 1, 1)

	p := pipeline.NewPipeline()
	p.Append(util.NewLineReader(*inputFile))
	p.Append(seq.NewFastAReader(1))
	p.Append(aligner)
	p.Append(seq.NewFastAWriter(*outputFile))

	p.Run()
}
