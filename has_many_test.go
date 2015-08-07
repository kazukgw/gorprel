package gorprel

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHasManyBuilder(t *testing.T) {
	dbmap, a := testInit(t)
	sb := dbmap.HasManyBuilder(&testGroup{1, "group", nil}, testUser{}, "*")
	sql, args, _ := sb.ToSql()
	a.Regexp("SELECT \\* FROM users WHERE group_id = \\?", sql)
	a.Equal([]interface{}{1}, args)
}

func TestHasMany(t *testing.T) {
	dbmap, a := testInit(t)
	cols := []string{"user_id", "name", "group_id"}
	sqlmock.ExpectQuery("SELECT \\* FROM users WHERE group_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).FromCSVString("1,John,1\n2,Taro,1\n3,Ben,1"))

	g := &testGroup{1, "group1", nil}
	_, err := dbmap.HasMany(g, testUser{})
	a.NotEmpty(g.TestUsers)
	a.Nil(err)
}
