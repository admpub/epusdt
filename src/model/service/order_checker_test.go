package service

import (
	"testing"
	"time"

	"github.com/admpub/pp"
	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	SetDefs([]*OrderCheckerDef{NewTronscanapiDef()})
	c := NewDefaultCheck(Defs())
	startTime := time.Now().Add(-24 * time.Hour).UnixMilli()
	endTime := time.Now().UnixMilli()
	def := Defs()[0]
	rows, err := c.query(def, `TVAz5k5NHtAXGwgKVjkX9xfiQuH8uiRs3q`, startTime, endTime)
	assert.NoError(t, err)
	pp.Println(rows)
	for _, row := range rows {
		result := def.ParseResult(row)
		amount, err := result.GetAmount(def)
		assert.NoError(t, err)
		assert.True(t, result.IsSuccess(def))
		pp.Println(amount, result)
		result.Release()
	}
}
