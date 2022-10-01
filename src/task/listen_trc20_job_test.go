package task

import (
	"fmt"
	"testing"

	"github.com/assimon/luuu/model/service"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestMarshalDefs(t *testing.T) {
	defs := []*service.OrderCheckerDef{
		service.NewTronscanapiDef(),
	}
	b, err := yaml.Marshal(defs)
	assert.NoError(t, err)
	fmt.Println(string(b))
	panic(err)
}
