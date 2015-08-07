package gorprel

import "gopkg.in/gorp.v1"

type DbMap struct {
	*gorp.DbMap
	Tracer
}

type Tracer interface {
	TraceOn()
	TraceOff()
}

func New(dbmap *gorp.DbMap, tracer Tracer) *DbMap {
	return &DbMap{
		DbMap:  dbmap,
		Tracer: tracer,
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
