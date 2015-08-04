package gorper

import (
	"database/sql"
	"errors"

	sq "github.com/lann/squirrel"
	"gopkg.in/gorp.v1"
)

type MappedToTable interface {
	SetTableMetas(tm *gorp.TableMap) *gorp.TableMap
	TableName() string
}

type Model interface {
	KeyName() string
	Key() interface{}
	IsNew() bool
	IsZero() bool
}

type MappedModel interface {
	MappedToTable
	Model
}

type Models []Model

func (d *DbMap) ToModels(rows []interface{}) Models {
	ms := make(Models, len(rows))
	for i, r := range rows {
		if m, ok := r.(Model); ok {
			ms[i] = m
		} else {
			panic("rows not implements Model interface")
		}
	}
	return ms
}

func (d *DbMap) ToRows(ms Models) []interface{} {
	rows := make([]interface{}, len(ms))
	for i, m := range ms {
		rows[i] = interface{}(m)
	}
	return rows
}

type HasOneAssociation interface {
	SetOneAssociation(Model)
}

type HasManyAssociation interface {
	SetManyAssociation(Models)
}

type Belongings interface {
	MappedModel
	FKName(MappedModel) string
	FK(MappedModel) interface{}
}

func (d *DbMap) BelongsToBuilder(m Belongings, belong MappedModel, selectStr string) sq.SelectBuilder {
	var slct string
	slct = selectStr
	if selectStr == "" {
		slct = "*"
	}

	t := belong.TableName()
	kname := m.FKName(belong)
	k := m.FK(belong)
	return sq.Select(slct).From(t).Where(sq.Eq{kname: k})
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
		if hoa, ok := m.(HasOneAssociation); ok {
			hoa.SetOneAssociation(ret)
		}
		return ret, nil
	}
	return nil, errors.New("model is not 'MappedModel'")
}

type HasMany interface {
	MappedModel
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
	if hma, ok := m.(HasManyAssociation); ok {
		hma.SetManyAssociation(ms)
	}
	return ms, nil
}

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
	if hma, ok := m.(HasManyAssociation); ok {
		hma.SetManyAssociation(ms)
	}
	if hoa, ok := m.(HasOneAssociation); ok {
		if len(ms) > 0 {
			hoa.SetOneAssociation(ms[0])
		}
	}
	return ms, nil
}

func (d *DbMap) Query(model interface{}, w sq.SelectBuilder) (Models, error) {
	q, args, err := w.ToSql()
	if err != nil {
		return Models{}, err
	}

	d.Tracer.TraceOn()
	rows, err := d.DbMap.Select(model, q, args...)
	d.Tracer.TraceOff()

	if len(rows) == 0 {
		err = sql.ErrNoRows
	}
	return d.ToModels(rows), err
}

func (d *DbMap) Get(holderHasKey MappedModel) error {
	q, args, err := sq.Select("*").From(holderHasKey.TableName()).
		Where(holderHasKey.KeyName()+" = ?", holderHasKey.Key()).ToSql()
	if err != nil {
		return err
	}
	d.Tracer.TraceOn()
	err = d.DbMap.SelectOne(holderHasKey, q, args...)
	d.Tracer.TraceOff()
	return err
}

func (d *DbMap) FindWhere(holder MappedModel, eq map[string]interface{}) error {
	q, args, err := sq.Select("*").From(holder.TableName()).
		Where(sq.Eq(eq)).ToSql()
	if err != nil {
		return err
	}
	d.Tracer.TraceOn()
	err = d.DbMap.SelectOne(holder, q, args...)
	d.Tracer.TraceOff()
	return err
}

func (d *DbMap) WhereBuilder(
	m MappedModel,
	eq map[string]interface{},
	selectStr string,
) sq.SelectBuilder {
	if selectStr == "" {
		selectStr = "*"
	}
	return sq.Select(selectStr).From(m.TableName()).Where(sq.Eq(eq))
}

func (d *DbMap) Where(m MappedModel, eq map[string]interface{}) (Models, error) {
	return d.Query(m, d.WhereBuilder(m, eq, ""))
}
