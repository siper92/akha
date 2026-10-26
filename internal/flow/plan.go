package flow

import (
	"errors"
	"fmt"
	"slices"
)

var _ Plan = (*Sequence)(nil)

type Sequence struct {
	order []ID
	index map[ID]int
	errs  []error
}

func NewPlan(ids ...ID) *Sequence {
	p := &Sequence{index: make(map[ID]int, len(ids))}
	for _, id := range ids {
		if _, ok := p.index[id]; ok {
			p.errs = append(p.errs, fmt.Errorf("%w: %s", ErrDuplicateID, id))
			continue
		}
		p.index[id] = len(p.order)
		p.order = append(p.order, id)
	}
	return p
}

func (p *Sequence) Order() []ID { return slices.Clone(p.order) }

func (p *Sequence) Index(id ID) (int, bool) {
	i, ok := p.index[id]
	return i, ok
}

func (p *Sequence) Validate() error { return errors.Join(p.errs...) }
