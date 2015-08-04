package gorper

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/lann/squirrel"
	"github.com/stretchr/testify/assert"
	"gopkg.in/gorp.v1"
)

func _connectToMock() (*gorp.DbMap, error) {
	db, err := sql.Open("mock", "")
	if err != nil {
		return &gorp.DbMap{}, err
	}

	return &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}, nil
}

func _init(t *testing.T) (*DbMap, *assert.Assertions) {
	a := assert.New(t)
	gorpdbmap, err := _connectToMock()
	if err != nil {
		t.Error(err)
	}
	return New(gorpdbmap, new(dummyTracer), dummyTablemaps), a
}

type dummyTracer struct{}

func (t *dummyTracer) TraceOn()  {}
func (t *dummyTracer) TraceOff() {}

type DummyGroup struct {
	GroupID int    `db:"group_id"`
	Name    string `db:"name"`

	Users *[]*DummyUser `db:"-"`
}

func (u DummyGroup) TableName() string {
	return "groups"
}
func (u DummyGroup) SetTableMetas(tm *gorp.TableMap) *gorp.TableMap {
	return tm.SetKeys(false, "group_id")
}
func (u DummyGroup) KeyName() string {
	return "group_id"
}
func (u DummyGroup) Key() interface{} {
	return u.GroupID
}
func (u DummyGroup) IsNew() bool {
	return u.GroupID == 0
}
func (u DummyGroup) IsZero() bool {
	return u == DummyGroup{}
}
func (g DummyGroup) FKNameInBelongings(MappedModel) string {
	return "group_id"
}
func (g DummyGroup) FKInBelongings(MappedModel) interface{} {
	return g.GroupID
}
func (g *DummyGroup) SetManyAssociation(ms Models) {
	if len(ms) == 0 {
		return
	}

	m := ms[0]
	if _, ok := m.(*DummyUser); ok {
		us := make([]*DummyUser, len(ms))
		for i, m := range ms {
			us[i] = m.(*DummyUser)
		}
		g.Users = &us
	}
}

type DummyUser struct {
	UserID  int    `db:"user_id"`
	Name    string `db:"name"`
	GroupID int    `db:"group_id"`

	*DummyGroup `db:"-"`
	Images      *[]*DummyImage `db:"-"`
}

func (u DummyUser) TableName() string {
	return "users"
}
func (u DummyUser) SetTableMetas(tm *gorp.TableMap) *gorp.TableMap {
	return tm.SetKeys(false, "user_id")
}
func (u DummyUser) KeyName() string {
	return "user_id"
}
func (u DummyUser) Key() interface{} {
	return u.UserID
}
func (u DummyUser) IsNew() bool {
	return u.UserID == 0
}
func (u DummyUser) IsZero() bool {
	return u == DummyUser{}
}
func (u DummyUser) FKName(MappedModel) string {
	return "group_id"
}
func (u DummyUser) FK(MappedModel) interface{} {
	return u.GroupID
}
func (u *DummyUser) SetOneAssociation(m Model) {
	if g, ok := m.(*DummyGroup); ok {
		u.DummyGroup = g
	}
}
func (u *DummyUser) SetManyAssociation(ms Models) {
	if len(ms) == 0 {
		return
	}

	m := ms[0]
	if _, ok := m.(*DummyImage); ok {
		imgs := make([]*DummyImage, len(ms))
		for i, m := range ms {
			imgs[i] = m.(*DummyImage)
		}
		u.Images = &imgs
	}
}

type DummyImage struct {
	ImageID int    `db:"image_id"`
	Name    string `db:"name"`
	URL     string `db:"url"`
}

func (u DummyImage) TableName() string {
	return "images"
}
func (u DummyImage) SetTableMetas(tm *gorp.TableMap) *gorp.TableMap {
	return tm.SetKeys(false, "image_id")
}
func (u DummyImage) KeyName() string {
	return "image_id"
}
func (u DummyImage) Key() interface{} {
	return u.ImageID
}
func (u DummyImage) IsNew() bool {
	return u.ImageID == 0
}
func (u DummyImage) IsZero() bool {
	return u == DummyImage{}
}

type DummyMapping struct {
	MappingID int `db:"mapping_id"`
	UserID    int `db:"user_id"`
	ImageID   int `db:"image_id"`
}

func (u DummyMapping) TableName() string {
	return "mappings"
}
func (u DummyMapping) SetTableMetas(tm *gorp.TableMap) *gorp.TableMap {
	return tm.SetKeys(false, "mapping_id")
}
func (u DummyMapping) KeyName() string {
	return "mapping_id"
}
func (u DummyMapping) Key() interface{} {
	return u.MappingID
}
func (u DummyMapping) IsNew() bool {
	return u.MappingID == 0
}
func (u DummyMapping) IsZero() bool {
	return u == DummyMapping{}
}
func (u DummyMapping) OtherModel(m MappedModel) MappedModel {
	switch m.TableName() {
	case "users":
		return DummyImage{}
	case "images":
		return DummyUser{}
	}
	panic("mapping model not has other model")
}
func (u DummyMapping) OtherKey(m MappedModel) interface{} {
	switch m.TableName() {
	case "users":
		return u.ImageID
	case "images":
		return u.UserID
	}
	panic("mapping model not has other model")
}

var dummyTablemaps = map[string]interface{}{
	"users":  DummyUser{},
	"images": DummyImage{},
}

func TestNew(t *testing.T) {
	a := assert.New(t)
	gorpdbmap, err := _connectToMock()
	if err != nil {
		t.Error(err)
		return
	}
	dbmap := New(gorpdbmap, new(dummyTracer), dummyTablemaps)
	a.IsType(dbmap, new(DbMap), "hogehgoe")
}

func TestCreate(t *testing.T) {
	dbmap, a := _init(t)

	sqlmock.ExpectExec("insert into `users` (.+)").
		WithArgs(1, "John", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := dbmap.Create(&DummyUser{1, "John", 1, nil, nil})
	a.Nil(err)
}

func TestUpdate(t *testing.T) {
	dbmap, a := _init(t)

	sqlmock.ExpectExec("update `users` (.+)").
		WithArgs(1, "John", 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	u := &DummyUser{1, "John", 1, nil, nil}
	_, err := dbmap.Update(u)
	a.Nil(err)
}

func TestDelete(t *testing.T) {
	dbmap, a := _init(t)

	sqlmock.ExpectExec("delete from `users` where (.+)").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := dbmap.Delete(&DummyUser{1, "John", 1, nil, nil})
	a.Nil(err)
}

func TestQuery(t *testing.T) {
	dbmap, a := _init(t)
	cols := []string{"user_id", "name", "group_id"}

	sqlmock.ExpectQuery("SELECT (.*) FROM users WHERE (.*)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).FromCSVString("1,John,1"))

	q := sq.Select("*").From("users").Where(sq.Eq{"user_id": 1})
	_, err := dbmap.Query(DummyUser{}, q)
	a.Nil(err)
}

func TestTransaction(t *testing.T) {
	dbmap, a := _init(t)

	sqlmock.ExpectBegin()
	sqlmock.ExpectRollback()

	err := dbmap.Transaction(func(tr *gorp.Transaction) error {
		return errors.New("")
	})
	a.Error(err, "DB should rollback when error returned from func that as transation arg.")
}

func TestBelongsToBuilder(t *testing.T) {
	dbmap, a := _init(t)
	sb := dbmap.BelongsToBuilder(DummyUser{1, "John", 1, nil, nil}, DummyGroup{1, "group1", nil}, "*")
	sql, args, _ := sb.ToSql()
	a.Regexp("SELECT \\* FROM groups WHERE group_id = \\?", sql)
	a.Equal([]interface{}{1}, args)
}

func TestGetBelongsTo(t *testing.T) {
	dbmap, a := _init(t)
	cols := []string{"group_id", "name"}
	sqlmock.ExpectQuery("SELECT \\* FROM groups WHERE group_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).FromCSVString("1,group1"))

	u := &DummyUser{1, "John", 1, nil, nil}
	_, err := dbmap.BelongsTo(u, DummyGroup{})
	a.NotNil(u.DummyGroup)
	a.Nil(err)
}

func TestHasManyBuilder(t *testing.T) {
	dbmap, a := _init(t)
	sb := dbmap.HasManyBuilder(DummyGroup{1, "group", nil}, DummyUser{}, "*")
	sql, args, _ := sb.ToSql()
	a.Regexp("SELECT \\* FROM users WHERE group_id = \\?", sql)
	a.Equal([]interface{}{1}, args)
}

func TestHasMany(t *testing.T) {
	dbmap, a := _init(t)
	cols := []string{"user_id", "name", "group_id"}
	sqlmock.ExpectQuery("SELECT \\* FROM users WHERE group_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).FromCSVString("1,John,1\n2,Taro,1\n3,Ben,1"))

	g := &DummyGroup{1, "group1", nil}
	_, err := dbmap.HasMany(g, DummyUser{})
	a.NotEmpty(g.Users)
	a.Nil(err)
}

func TestManyToManyBuilder(t *testing.T) {
	dbmap, a := _init(t)

	cols := []string{"mapping_id", "user_id", "image_id"}
	sqlmock.ExpectQuery("SELECT \\* FROM mappings WHERE user_id = (.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).FromCSVString("1,1,1"))

	sb, err := dbmap.ManyToManyBuilder(&DummyUser{UserID: 1}, DummyMapping{}, "*")
	a.Nil(err)
	sql, args, _ := sb.ToSql()

	a.Regexp("SELECT \\* FROM images WHERE image_id IN \\(\\?\\)", sql)
	a.Equal([]interface{}{1}, args)
}

func TestGetOthersByMapping(t *testing.T) {
	dbmap, a := _init(t)

	cols := []string{"mapping_id", "user_id", "image_id"}
	sqlmock.ExpectQuery("SELECT \\* FROM mappings WHERE user_id = (.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).FromCSVString("1,1,1\n1,1,2\n1,1,3"))

	imgcols := []string{"image_id", "name", "url"}
	sqlmock.ExpectQuery("SELECT \\* FROM images WHERE image_id IN (.+)").
		WithArgs(1, 2, 3).
		WillReturnRows(sqlmock.NewRows(imgcols).FromCSVString(
		`
1,name1,http://1
2,name2,http://2
3,name3,http://3
		`))

	u := &DummyUser{UserID: 1}
	_, err := dbmap.ManyToMany(u, DummyMapping{})
	a.Nil(err)
	a.NotNil(u.Images)
	a.NotEmpty(u.Images)
}

func TestGet(t *testing.T) {
	dbmap, a := _init(t)
	cols := []string{"user_id", "name", "group_id"}
	sqlmock.ExpectQuery("SELECT \\* FROM users WHERE user_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).FromCSVString("1,John,1"))

	err := dbmap.Get(&DummyUser{UserID: 1})
	a.Nil(err)
}

func TestWhereBuilder(t *testing.T) {
	dbmap, a := _init(t)
	sb := dbmap.WhereBuilder(DummyUser{}, map[string]interface{}{"user_id": 1}, "*")
	sql, args, _ := sb.ToSql()
	a.Regexp("SELECT \\* FROM users WHERE user_id = (.+)", sql)
	a.Equal([]interface{}{1}, args)
}
