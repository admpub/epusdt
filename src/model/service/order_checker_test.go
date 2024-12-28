package service

import (
	"testing"

	"github.com/admpub/pp"
	"github.com/dromara/carbon/v2"
	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	SetDefs([]*OrderCheckerDef{NewTronscanapiDef()})
	c := NewDefaultCheck(Defs())
	startTime := carbon.Now().AddHours(-24).TimestampMilli()
	endTime := carbon.Now().TimestampMilli()
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
