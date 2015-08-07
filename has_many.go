package gorprel

import sq "github.com/lann/squirrel"

type HasManySetter interface {
	SetHasMany(Models)
}

type HasMany interface {
	MappedModel
	HasManySetter
	FKNameInBelongings(MappedModel) string
	FKInBelongings(MappedModel) interface{}
}

func (d *DbMap) HasManyBuilder(m HasMany, b MappedModel, selectStr string) sq.SelectBuilder {
	var slct string
	slct = selectStr
	if selectStr == "" {
		slct = "*"
	}

	kname := m.FKNameInBelongings(b)
	return sq.Select(slct).From(b.TableName()).Where(sq.Eq{kname: m.FKInBelongings(b)})
}

func (d *DbMap) HasMany(m HasMany, b MappedModel) (Models, error) {
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
