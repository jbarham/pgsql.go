package pgsql

import (
	"github.com/bmizerany/assert"
	"http"
	"os"
	"testing"
)

func TestURL(t *testing.T) {
	cs, err := url2params("postgres://example.com")
	assert.Equal(t, ErrMissingDbName, err)
	assert.Equal(t, "", cs)

	cs, err = url2params("")
	exp := &http.URLError{"parse", "", os.NewError("empty url")}
	assert.Equal(t, exp, err)
	assert.Equal(t, "", cs)

	cs, err = url2params("postgres://blake@example.com/mydb")
	assert.Equal(t, nil, err)
	assert.Equal(t, "dbname=mydb user=blake host=example.com", cs)

	cs, err = url2params("postgres://:secret@example.com/mydb")
	assert.Equal(t, nil, err)
	assert.Equal(t, "dbname=mydb password=secret host=example.com", cs)

	cs, err = url2params("postgres://:secret@example.com/mydb")
	assert.Equal(t, nil, err)
	assert.Equal(t, "dbname=mydb password=secret host=example.com", cs)

	cs, err = url2params("postgres://@example.com:1234/mydb")
	assert.Equal(t, nil, err)
	assert.Equal(t, "dbname=mydb host=example.com port=1234", cs)
}
