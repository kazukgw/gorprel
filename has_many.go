package gorprel

import sq "github.com/Masterminds/squirrel"

type HasManySetter interface {
	SetHasMany(Models)
}

type HasMany interface {
	Model
	HasManySetter
	FKNameInBelongings(Model) string
	FKInBelongings(Model) interface{}
}

func (d *DbMap) HasManyBuilder(m HasMany, b Model, selectStr string) sq.SelectBuilder {
	var slct string
	slct = selectStr
	if selectStr == "" {
		slct = "*"
	}

	kname := m.FKNameInBelongings(b)
	return sq.Select(slct).From(b.TableName()).Where(sq.Eq{kname: m.FKInBelongings(b)})
}

func (d *DbMap) HasMany(m HasMany, b Model) (Models, error) {
	sb := d.HasManyBuilder(m, b, "")
	ms, err := d.Query(b, sb)
	if err != nil {
		return ms, err
	}
	if hma, ok := m.(HasManySetter); ok {
		hma.SetHasMany(ms)
	}
	return ms, nil
}
