package gorprel

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestManyToManyBuilder(t *testing.T) {
	dbmap, a := testInit(t)

	cols := []string{"mapping_id", "tag_id", "image_id"}
	sqlmock.ExpectQuery("SELECT \\* FROM mappings WHERE tag_id = (.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).FromCSVString("1,1,1"))

	sb, err := dbmap.ManyToManyBuilder(&testTag{TagID: 1}, testMapping{}, "*")
	a.Nil(err)
	sql, args, _ := sb.ToSql()

	a.Regexp("SELECT \\* FROM images WHERE image_id IN \\(\\?\\)", sql)
	a.Equal([]interface{}{1}, args)
}

func TestGetOthersByMapping(t *testing.T) {
	dbmap, a := testInit(t)

	cols := []string{"mapping_id", "tag_id", "image_id"}
	sqlmock.ExpectQuery("SELECT \\* FROM mappings WHERE tag_id = (.+)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).FromCSVString("1,1,1\n1,1,2\n1,1,3"))

	imgcols := []string{"image_id", "name", "url"}
	sqlmock.ExpectQuery("SELECT \\* FROM images WHERE image_id IN (.+)").
		WithArgs(1, 2, 3).
		WillReturnRows(sqlmock.NewRows(imgcols).FromCSVString(
		` 1,name1,http://1
          2,name2,http://2
          3,name3,http://3
		`))

	tag := &testTag{TagID: 1}
	_, err := dbmap.ManyToMany(tag, testMapping{})
	a.Nil(err)
	a.NotNil(tag.TestImages)
	a.NotEmpty(tag.TestImages)
}
