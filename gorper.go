package gorper

import (
	sq "github.com/lann/squirrel"
	"gopkg.in/gorp.v1"
)

type DbMap struct {
	*gorp.DbMap
	Tracer
	TableMaps map[string]interface{}
}

type Tracer interface {
	TraceOn()
	TraceOff()
}

func New(dbmap *gorp.DbMap, tracer Tracer, tablemaps map[string]interface{}) *DbMap {
	for k, m := range tablemaps {
		tm := dbmap.AddTableWithName(m, k)
		tms, ok := m.(MappedToTable)
		if ok {
			_ = tms.SetTableMetas(tm)
		}
	}
	return &DbMap{
		DbMap:     dbmap,
		Tracer:    tracer,
		TableMaps: tablemaps,
	}
}

func (d *DbMap) Create(list ...interface{}) error {
	d.Tracer.TraceOn()
	err := d.DbMap.Insert(list...)
	d.Tracer.TraceOff()
	return err
}

func (d *DbMap) Update(list ...interface{}) (int64, error) {
	d.Tracer.TraceOn()
	i, err := d.DbMap.Update(list...)
	d.Tracer.TraceOff()
	return i, err
}

func (d *DbMap) Save(m Model) error {
	if m.IsNew() {
		return d.Create(m)
	}
	_, err := d.Update(m)
	return err
}

func (d *DbMap) FindOrCreate(holder MappedModel, key interface{}) error {
	d.Tracer.TraceOn()
	q := "select * from " + holder.TableName() + " where " + holder.KeyName() + " = ?;"
	err := d.DbMap.SelectOne(holder, q, key)
	d.Tracer.TraceOff()
	if err != nil {
		return d.Create(holder)
	}
	return nil
}

func (d *DbMap) Delete(list ...interface{}) (int64, error) {
	d.Tracer.TraceOn()
	i, err := d.DbMap.Delete(list...)
	d.Tracer.TraceOff()
	return i, err
}

func (d *DbMap) Transaction(fn func(tr *gorp.Transaction) error) error {
	tr, err := d.DbMap.Begin()
	if err != nil {
		return err
	}
	err = fn(tr)
	if err != nil {
		if rerr := tr.Rollback(); rerr != nil {
			panic(rerr.Error())
		}
		return err
	}
	if cerr := tr.Commit(); cerr != nil {
		panic(cerr.Error())
	}
	return nil
}

func (d *DbMap) Exists(m MappedModel) bool {
	table := m.TableName()
	keyname := m.KeyName()
	key := m.Key()
	count, err := d.SelectInt("select count(*) from "+table+" where "+keyname+" = ?", key)
	if err != nil {
		return false
	}
	return count > 0
}

func (d *DbMap) CountBuilder(m MappedModel, eq map[string]interface{}) sq.SelectBuilder {
	return sq.Select("count(*)").From(m.TableName()).Where(sq.Eq(eq))
}

func (d *DbMap) Count(m MappedModel, eq map[string]interface{}) (int, error) {
	sb := d.CountBuilder(m, eq)
	q, args, err := sb.ToSql()
	if err != nil {
		return 0, err
	}
	count, err := d.SelectInt(q, args...)
	return int(count), nil
}
