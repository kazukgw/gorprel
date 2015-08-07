package gorprel

import (
	"errors"

	sq "github.com/lann/squirrel"
)

type HasOneSetter interface {
	SetHasOne(Model)
}

type HasOne interface {
	MappedModel
	HasOneSetter
	KeyNameInOne(MappedModel) string
	KeyInOne(MappedModel) interface{}
}

func (d *DbMap) HasOneBuilder(m HasOne, one MappedModel, selectStr string) sq.SelectBuilder {
	if selectStr == "" {
		selectStr = "*"
	}
	t := one.TableName()
	kname := m.KeyNameInOne(one)
	k := m.KeyInOne(one)
	return sq.Select(selectStr).From(t).Where(sq.Eq{kname: k})
}

func (d *DbMap) HasOne(m HasOne, one MappedModel) (MappedModel, error) {
	ms, err := d.Query(m, d.HasOneBuilder(m, one, ""))
	if err != nil {
		return nil, err
	}
	if len(ms) == 0 {
		return nil, errors.New("Model is not found")
	}

	if ret, ok := ms[0].(MappedModel); ok {
		if hoa, ok := m.(HasOneSetter); ok {
			hoa.SetHasOne(ret)
		}
		return ret, nil
	}
	return nil, errors.New("model is not 'MappedModel'")
}
