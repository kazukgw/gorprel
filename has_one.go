package gorprel

import sq "github.com/Masterminds/squirrel"

type HasOneSetter interface {
	SetHasOne(Model)
}

type HasOne interface {
	Model
	HasOneSetter
	FKName(Model) string
	FK(Model) interface{}
}

func (d *DbMap) HasOneBuilder(m HasOne, theOne Model, selectStr string) sq.SelectBuilder {
	if selectStr == "" {
		selectStr = "*"
	}
	t := d.TableName(theOne)
	kname := m.FKName(theOne)
	k := m.FK(theOne)
	return sq.Select(selectStr).From(t).Where(sq.Eq{kname: k})
}

func (d *DbMap) HasOne(m HasOne, one Model) (Model, error) {
	ms, err := d.Query(one, d.HasOneBuilder(m, one, ""))
	if err != nil {
		return nil, err
	}
	mm := ms[0].(Model)
	m.SetHasOne(mm)
	return mm, nil
}
