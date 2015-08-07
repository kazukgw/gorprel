package gorprel

import sq "github.com/lann/squirrel"

type Mapping interface {
	MappedModel
	OtherModel(MappedModel) MappedModel
	OtherKey(MappedModel) interface{}
}

func (d *DbMap) ManyToManyBuilder(m MappedModel, mapping Mapping, selectStr string) (sq.SelectBuilder, error) {
	var slct string
	slct = selectStr
	if selectStr == "" {
		slct = "*"
	}
	other := mapping.OtherModel(m)
	kname := m.KeyName()
	k := m.Key()
	w := sq.Select("*").From(mapping.TableName()).Where(sq.Eq{kname: k})

	rows, err := d.Query(mapping, w)
	if err != nil {
		return sq.SelectBuilder{}, err
	}

	keys := make([]interface{}, 0)
	for _, r := range rows {
		if itr, ok := r.(Mapping); ok {
			keys = append(keys, itr.OtherKey(m))
		}
	}

	kname = other.KeyName()
	return sq.Select(slct).From(other.TableName()).Where(sq.Eq{kname: keys}), nil
}

func (d *DbMap) ManyToMany(m MappedModel, mapping Mapping) (Models, error) {
	sb, err := d.ManyToManyBuilder(m, mapping, "")
	if err != nil {
		return Models{}, err
	}
	ms, err := d.Query(mapping.OtherModel(m), sb)
	if err != nil {
		return ms, err
	}
	if hma, ok := m.(HasManySetter); ok {
		hma.SetHasMany(ms)
	}
	if hoa, ok := m.(HasOneSetter); ok {
		if len(ms) > 0 {
			hoa.SetHasOne(ms[0])
		}
	}
	return ms, nil
}
