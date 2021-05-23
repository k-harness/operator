package variables

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTemplateFunctions(t *testing.T) {
	rand.Seed(time.Now().Unix())

	x := New(nil, nil)

	t.Run("uuid", func(t *testing.T) {
		body := `{{ uuid }}`
		res := x.Template(body)

		_, err := uuid.Parse(res)
		assert.NoError(t, err)
	})

	t.Run("rnd_str", func(t *testing.T) {
		const checkLen = 10

		body := spew.Sprintf(`{{ rnd_str %d}}`, checkLen)
		res := x.Template(body)

		assert.Len(t, res, checkLen)
	})

	t.Run("range_int", func(t *testing.T) {
		test := []struct {
			name     string
			min, max int
			err      bool
		}{
			{"min > max", 2, 1, true},
			{"min == max", 2, 2, true},
			{"min < max", 1, 2, false},
		}

		for _, v := range test {
			t.Run(v.name, func(t *testing.T) {
				body := spew.Sprintf(`{{ range_int %d %d}}`, v.min, v.max)
				res, err := x.TemplateBytes([]byte(body))

				assert.Equal(t, v.err, err != nil, err)
				fmt.Println(string(res))
			})
		}
	})
	t.Run("conditions", func(t *testing.T) {
		const body = `{{$rv := range_int 1 100}}{{ $ch := le $rv 30}}{{ if $ch }}{{ .WIN }}{{else }}{{ .LOSE }}{{end}}`
		for i := 0; i < 10; i++ {
			x.Update("WIN", "10.00")
			x.Update("LOSE", "0.00")
			res, err := x.TemplateBytes([]byte(body))

			assert.NoError(t, err)
			fmt.Println(string(res))
		}
	})
}
