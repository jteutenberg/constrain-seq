# constrain-seq
Modify sequences to match arbitrary constraints using the minimal number of insertion and deletion edits.

This is a general library for normalising sequences against constraints. 

An example compilable command "partial-align" applies the following constraints:
- the sequence is over the alphabet {A,C,G,T}
- the sequence will have no homopolymers (repeated bases)
- the sequence will match the beginning of a given target partial sequence (with 'N' wild-card bases)
- the sequence will not extend beyong a given maximum length
