package checker

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestTemplateFunctions(t *testing.T) {
	rand.Seed(time.Now().Unix())

	t.Run("uuid", func(t *testing.T) {
		body := `{{ uuid }}`
		b := Body(&v1alpha1.Body{Row: body})

		res, err := b.GetBody(nil)
		assert.NoError(t, err)
		_, err = uuid.Parse(string(res))
		assert.NoError(t, err)
	})

	t.Run("rnd_str", func(t *testing.T) {
		const checkLen = 10

		body := spew.Sprintf(`{{ rnd_str %d}}`, checkLen)
		b := Body(&v1alpha1.Body{Row: body})

		res, err := b.GetBody(nil)
		assert.NoError(t, err)
		assert.Len(t, string(res), checkLen)
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
				b := Body(&v1alpha1.Body{Row: body})
				res, err := b.GetBody(nil)
				assert.Equal(t, v.err, err != nil, err)
				fmt.Println(string(res))
			})
		}
	})
	t.Run("conditions", func(t *testing.T) {
		const body = `{{$rv := range_int 1 100}}{{ $ch := le $rv 30}}{{ if $ch }}{{ .WIN }}{{else }}{{ .LOSE }}{{end}}`
		for i := 0; i < 10; i++ {
			b := Body(&v1alpha1.Body{Row: body})
			res, err := b.GetBody(map[string]string{"WIN": "10.00", "LOSE": "0.00"})
			assert.NoError(t, err)
			fmt.Println(string(res))
		}
	})
}
