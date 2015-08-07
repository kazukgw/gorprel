package gorprel

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestBelongsToBuilder(t *testing.T) {
	dbmap, a := testInit(t)
	sb := dbmap.BelongsToBuilder(&testUser{1, "John", 1, nil, nil}, testGroup{}, "")
	sql, args, _ := sb.ToSql()
	a.Regexp("SELECT \\* FROM groups WHERE group_id = \\?", sql)
	a.Equal([]interface{}{1}, args)
}

func TestBelongsTo(t *testing.T) {
	dbmap, a := testInit(t)
	cols := []string{"group_id", "name"}
	sqlmock.ExpectQuery("SELECT \\* FROM groups WHERE group_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).FromCSVString("1,group1"))

	u := &testUser{1, "John", 1, nil, nil}
	_, err := dbmap.BelongsTo(u, testGroup{})
	a.NotNil(u.TestGroup)
	a.Nil(err)
}
