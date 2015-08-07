package gorprel

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/lann/squirrel"
)

func TestGet(t *testing.T) {
	dbmap, a := testInit(t)
	cols := []string{"user_id", "name", "group_id"}
	sqlmock.ExpectQuery("SELECT \\* FROM users WHERE user_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).FromCSVString("1,John,1"))

	err := dbmap.Get(&testUser{UserID: 1})
	a.Nil(err)
}

func TestWhereBuilder(t *testing.T) {
	dbmap, a := testInit(t)
	sb := dbmap.WhereBuilder(testUser{}, map[string]interface{}{"user_id": 1}, "*")
	sql, args, _ := sb.ToSql()
	a.Regexp("SELECT \\* FROM users WHERE user_id = (.+)", sql)
	a.Equal([]interface{}{1}, args)
}

func TestQuery(t *testing.T) {
	dbmap, a := testInit(t)
	cols := []string{"user_id", "name", "group_id"}

	sqlmock.ExpectQuery("SELECT (.*) FROM users WHERE (.*)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).FromCSVString("1,John,1"))

	q := sq.Select("*").From("users").Where(sq.Eq{"user_id": 1})
	_, err := dbmap.Query(testUser{}, q)
	a.Nil(err)
}
