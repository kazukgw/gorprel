package gorprel

import (
	"errors"

	sq "github.com/lann/squirrel"
)

type BelongsToSetter interface {
	SetBelongsTo(Model)
}

type Belongings interface {
	MappedModel
	BelongsToSetter
	FKName(MappedModel) string
	FK(MappedModel) interface{}
}

func (d *DbMap) BelongsToBuilder(m Belongings, belong MappedModel, selectStr string) sq.SelectBuilder {
	if selectStr == "" {
		selectStr = "*"
	}
	t := belong.TableName()
	kname := m.FKName(belong)
	k := m.FK(belong)
	return sq.Select(selectStr).From(t).Where(sq.Eq{kname: k})
}

func (d *DbMap) BelongsTo(m Belongings, belong MappedModel) (MappedModel, error) {
	ms, err := d.Query(belong, d.BelongsToBuilder(m, belong, ""))
	if err != nil {
		return nil, err
	}
	mm := ms[0].(MappedModel)
	m.SetBelongsTo(mm)
	return mm, nil
}
