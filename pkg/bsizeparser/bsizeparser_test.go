package bsizeparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	var (
		r   uint64
		err error
	)

	r, err = Parse("69")
	assert.Nil(t, err)
	assert.EqualValues(t, 69, r)

	r, err = Parse("  \t69 ")
	assert.Nil(t, err)
	assert.EqualValues(t, 69, r)

	r, err = Parse("69K")
	assert.Nil(t, err)
	assert.EqualValues(t, 69*1024, r)

	r, err = Parse("69k")
	assert.Nil(t, err)
	assert.EqualValues(t, 69*1024, r)

	r, err = Parse("    69K \t ")
	assert.Nil(t, err)
	assert.EqualValues(t, 69*1024, r)

	r, err = Parse("69m")
	assert.Nil(t, err)
	assert.EqualValues(t, 69*1024*1024, r)

	r, err = Parse("69M")
	assert.Nil(t, err)
	assert.EqualValues(t, 69*1024*1024, r)

	r, err = Parse("69g")
	assert.Nil(t, err)
	assert.EqualValues(t, 69*1024*1024*1024, r)

	r, err = Parse("69G")
	assert.Nil(t, err)
	assert.EqualValues(t, 69*1024*1024*1024, r)

	r, err = Parse("69t")
	assert.Nil(t, err)
	assert.EqualValues(t, 69*1024*1024*1024*1024, r)

	r, err = Parse("69T")
	assert.Nil(t, err)
	assert.EqualValues(t, 69*1024*1024*1024*1024, r)

	r, err = Parse("69p")
	assert.Nil(t, err)
	assert.EqualValues(t, 69*1024*1024*1024*1024*1024, r)

	r, err = Parse("69P")
	assert.Nil(t, err)
	assert.EqualValues(t, 69*1024*1024*1024*1024*1024, r)

	r, err = Parse("M")
	assert.Nil(t, err)
	assert.EqualValues(t, 1024*1024, r)

	r, err = Parse("")
	assert.Nil(t, err)
	assert.EqualValues(t, 0, r)

	r, err = Parse("sdfm")
	assert.NotNil(t, err)
	assert.EqualValues(t, 0, r)
}
