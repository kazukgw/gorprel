package gorprel

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHasOneBuilder(t *testing.T) {
	dbmap, a := testInit(t)
	sb := dbmap.HasOneBuilder(&testImage{ImageID: 1, AuthorID: 1}, testAuthor{}, "")
	sql, args, _ := sb.ToSql()
	a.Regexp("SELECT \\* FROM authors WHERE author_id = \\?", sql)
	a.Equal([]interface{}{1}, args)
}

func TestHasOne(t *testing.T) {
	dbmap, a := testInit(t)
	cols := []string{"author_id", "name"}
	sqlmock.ExpectQuery("SELECT \\* FROM authors WHERE author_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).FromCSVString("1,John"))

	i := &testImage{ImageID: 1, Name: "image1", URL: "http://image1", AuthorID: 1}
	_, err := dbmap.HasOne(i, testAuthor{})
	a.NotNil(i.TestAuthor)
	a.Nil(err)
}
