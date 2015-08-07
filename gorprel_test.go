package gorprel

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/gorp.v1"
)

func connectToMock() (*gorp.DbMap, error) {
	db, err := sql.Open("mock", "")
	if err != nil {
		return &gorp.DbMap{}, err
	}

	return &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}, nil
}

var testTablemaps = map[string]interface{}{
	"groups":   testGroup{},
	"users":    testUser{},
	"images":   testImage{},
	"tags":     testTag{},
	"mappings": testMapping{},
	"author":   testAuthor{},
}

type TableOptionSetter interface {
	SetTableOptions(tm *gorp.TableMap) *gorp.TableMap
}

func testInit(t *testing.T) (*DbMap, *assert.Assertions) {
	a := assert.New(t)
	gorpdbmap, err := connectToMock()
	for k, m := range testTablemaps {
		tm := gorpdbmap.AddTableWithName(m, k)
		tms, ok := m.(TableOptionSetter)
		if ok {
			_ = tms.SetTableOptions(tm)
		}
	}
	if err != nil {
		t.Error(err)
	}
	return New(gorpdbmap, new(testTracer)), a
}

type testTracer struct{}

func (t *testTracer) TraceOn()  {}
func (t *testTracer) TraceOff() {}

// testGroup implements HasMany interface
type testGroup struct {
	GroupID int    `db:"group_id"`
	Name    string `db:"name"`

	TestUsers *[]*testUser `db:"-"`
}

func (g testGroup) TableName() string {
	return "groups"
}
func (g testGroup) SetTableOptions(tm *gorp.TableMap) *gorp.TableMap {
	return tm.SetKeys(false, "group_id")
}
func (g testGroup) KeyName() string {
	return "group_id"
}
func (g testGroup) Key() interface{} {
	return g.GroupID
}
func (g testGroup) FKNameInBelongings(Model) string {
	return "group_id"
}
func (g testGroup) FKInBelongings(Model) interface{} {
	return g.GroupID
}
func (g *testGroup) SetHasMany(ms Models) {
	if len(ms) == 0 {
		return
	}

	m := ms[0]
	if _, ok := m.(*testUser); ok {
		us := make([]*testUser, len(ms))
		for i, m := range ms {
			us[i] = m.(*testUser)
		}
		g.TestUsers = &us
	}
}

// TestUser implements BelongsTo interface and HasMany interface (has many images)
type testUser struct {
	UserID  int    `db:"user_id"`
	Name    string `db:"name"`
	GroupID int    `db:"group_id"`

	TestGroup  *testGroup    `db:"-"`
	TestImages *[]*testImage `db:"-"`
}

func (u testUser) TableName() string {
	return "users"
}
func (u testUser) SetTableOptions(tm *gorp.TableMap) *gorp.TableMap {
	return tm.SetKeys(false, "user_id")
}
func (u testUser) KeyName() string {
	return "user_id"
}
func (u testUser) Key() interface{} {
	return u.UserID
}
func (u testUser) FKName(Model) string {
	return "group_id"
}
func (u testUser) FK(Model) interface{} {
	return u.GroupID
}
func (u *testUser) SetBelongsTo(m Model) {
	if g, ok := m.(*testGroup); ok {
		u.TestGroup = g
	}
}
func (u *testUser) SetHasMany(ms Models) {
	if len(ms) == 0 {
		return
	}

	m := ms[0]
	if _, ok := m.(*testImage); ok {
		imgs := make([]*testImage, len(ms))
		for i, m := range ms {
			imgs[i] = m.(*testImage)
		}
		u.TestImages = &imgs
	}
}

// TestImage belongs to TestUsers and is related to TestTag through TestMapping
// and implements HasOne interface
type testImage struct {
	ImageID    int         `db:"image_id"`
	Name       string      `db:"name"`
	URL        string      `db:"url"`
	AuthorID   int         `db:"author_id"`
	TestTags   *[]*testTag `db:"-"`
	TestAuthor *testAuthor `db:"-"`
}

func (i testImage) TableName() string {
	return "images"
}
func (i testImage) SetTableOptions(tm *gorp.TableMap) *gorp.TableMap {
	return tm.SetKeys(false, "image_id")
}
func (i testImage) KeyName() string {
	return "image_id"
}
func (i testImage) Key() interface{} {
	return i.ImageID
}
func (i testImage) FKName(m Model) string {
	return "author_id"
}
func (i testImage) FK(m Model) interface{} {
	return i.AuthorID
}
func (i *testImage) SetHasOne(m Model) {
	i.TestAuthor = m.(*testAuthor)
}

// TestTag is related to TestImage through TestMapping
type testTag struct {
	TagID      int           `db:"tag_id"`
	Name       string        `db:"name"`
	TestImages *[]*testImage `db:"-"`
}

func (tag testTag) TableName() string {
	return "tags"
}
func (tag testTag) SetTableOptions(tm *gorp.TableMap) *gorp.TableMap {
	return tm.SetKeys(false, "tag_id")
}
func (tag testTag) KeyName() string {
	return "tag_id"
}
func (tag testTag) Key() interface{} {
	return tag.TagID
}
func (tag *testTag) SetHasMany(ms Models) {
	images := make([]*testImage, 0)
	for _, m := range ms {
		if image, ok := m.(*testImage); ok {
			images = append(images, image)
		}
	}
	tag.TestImages = &images
}

// TestMapping implements Mapping interface
type testMapping struct {
	MappingID int `db:"mapping_id"`
	TagID     int `db:"tag_id"`
	ImageID   int `db:"image_id"`
}

func (mp testMapping) TableName() string {
	return "mappings"
}
func (mp testMapping) SetTableOptions(tm *gorp.TableMap) *gorp.TableMap {
	return tm.SetKeys(false, "mapping_id")
}
func (mp testMapping) KeyName() string {
	return "mapping_id"
}
func (mp testMapping) Key() interface{} {
	return mp.MappingID
}
func (mp testMapping) OtherModel(m Model) Model {
	switch m.TableName() {
	case "tags":
		return testImage{}
	case "images":
		return testTag{}
	}
	panic("mapping model not has other model")
}
func (mp testMapping) OtherKey(m Model) interface{} {
	switch m.TableName() {
	case "tags":
		return mp.ImageID
	case "images":
		return mp.TagID
	}
	panic("mapping model not has other model")
}

type testAuthor struct {
	AuthorID int    `db:"author_id"`
	Name     string `db:"name"`
}

func (a testAuthor) TableName() string {
	return "authors"
}
func (a testAuthor) SetTableOptions(tm *gorp.TableMap) *gorp.TableMap {
	return tm.SetKeys(false, "author_id")
}
func (a testAuthor) KeyName() string {
	return "author_id"
}
func (a testAuthor) Key() interface{} {
	return a.AuthorID
}

func TestNew(t *testing.T) {
	a := assert.New(t)
	gorpdbmap, err := connectToMock()
	if err != nil {
		t.Error(err)
		return
	}
	dbmap := New(gorpdbmap, new(testTracer))
	a.IsType(dbmap, new(DbMap), "")
}

func TestCreate(t *testing.T) {
	dbmap, a := testInit(t)

	sqlmock.ExpectExec("insert into `users` (.+)").
		WithArgs(1, "John", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := dbmap.Create(&testUser{1, "John", 1, nil, nil})
	a.Nil(err)
}

func TestUpdate(t *testing.T) {
	dbmap, a := testInit(t)

	sqlmock.ExpectExec("update `users` (.+)").
		WithArgs(1, "John", 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	u := &testUser{1, "John", 1, nil, nil}
	_, err := dbmap.Update(u)
	a.Nil(err)
}

func TestDelete(t *testing.T) {
	dbmap, a := testInit(t)

	sqlmock.ExpectExec("delete from `users` where (.+)").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := dbmap.Delete(&testUser{1, "John", 1, nil, nil})
	a.Nil(err)
}

func TestTransaction(t *testing.T) {
	dbmap, a := testInit(t)

	sqlmock.ExpectBegin()
	sqlmock.ExpectRollback()

	err := dbmap.Transaction(func(tr *gorp.Transaction) error {
		return errors.New("")
	})
	a.Error(err, "DB should rollback when error returned from func that as transation arg.")
}
