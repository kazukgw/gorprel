package gorprel

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
)

type Model interface {
	TableName() string
	KeyName() string
	Key() interface{}
}

type Models []Model

func (d *DbMap) ToModels(rows []interface{}) Models {
	ms := make(Models, len(rows))
	for i, r := range rows {
		if m, ok := r.(Model); ok {
			ms[i] = m
		} else {
			panic("rows not implements model interface")
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

func (d *DbMap) FindOrCreate(holder Model, key interface{}) error {
	d.Tracer.TraceOn()
	q := "select * from " + holder.TableName() + " where " + holder.KeyName() + " = ?;"
	err := d.DbMap.SelectOne(holder, q, key)
	d.Tracer.TraceOff()
	if err != nil {
		return d.Create(holder)
	}
	return nil
}

func (d *DbMap) Exists(m Model) bool {
	table := m.TableName()
	keyname := m.KeyName()
	key := m.Key()
	count, err := d.SelectInt("select count(*) from "+table+" where "+keyname+" = ?", key)
	if err != nil {
		return false
	}
	return count > 0
}

func (d *DbMap) CountBuilder(m Model, eq map[string]interface{}) sq.SelectBuilder {
	return sq.Select("count(*)").From(m.TableName()).Where(sq.Eq(eq))
}

func (d *DbMap) Count(m Model, eq map[string]interface{}) (int, error) {
	sb := d.CountBuilder(m, eq)
	q, args, err := sb.ToSql()
	if err != nil {
		return 0, err
	}
	count, err := d.SelectInt(q, args...)
	return int(count), nil
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

func (d *DbMap) Get(holderHasKey Model) error {
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

func (d *DbMap) FindWhere(holder Model, eq map[string]interface{}) error {
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
	m Model,
	eq map[string]interface{},
	selectStr string,
) sq.SelectBuilder {
	if selectStr == "" {
		selectStr = "*"
	}
	return sq.Select(selectStr).From(m.TableName()).Where(sq.Eq(eq))
}

func (d *DbMap) Where(m Model, eq map[string]interface{}) (Models, error) {
	return d.Query(m, d.WhereBuilder(m, eq, ""))
}
