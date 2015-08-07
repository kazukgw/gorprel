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
	if len(ms) == 0 {
		return nil, errors.New("Model is not found")
	}

	if ret, ok := ms[0].(MappedModel); ok {
		m.SetBelongsTo(ret)
		return ret, nil
	}
	return nil, errors.New("model is not 'MappedModel'")
}
