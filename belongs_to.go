package gorprel

import sq "github.com/Masterminds/squirrel"

type BelongsToSetter interface {
	SetBelongsTo(Model)
}

type Belongings interface {
	Model
	BelongsToSetter
	FKName(Model) string
	FK(Model) interface{}
}

func (d *DbMap) BelongsToBuilder(m Belongings, belong Model, selectStr string) sq.SelectBuilder {
	if selectStr == "" {
		selectStr = "*"
	}
	t := belong.TableName()
	kname := m.FKName(belong)
	k := m.FK(belong)
	return sq.Select(selectStr).From(t).Where(sq.Eq{kname: k})
}

func (d *DbMap) BelongsTo(m Belongings, belong Model) (Model, error) {
	ms, err := d.Query(belong, d.BelongsToBuilder(m, belong, ""))
	if err != nil {
		return nil, err
	}
	mm := ms[0].(Model)
	m.SetBelongsTo(mm)
	return mm, nil
}
