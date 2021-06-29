package variables

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTemplateFunctions(t *testing.T) {
	rand.Seed(time.Now().Unix())

	x := New()

	t.Run("func execution for variable inside", func(t *testing.T) {
		xx := New(map[string]string{
			"TIME": `{{ unix }}`,
		})

		body := `{{ .TIME}}`
		res := xx.Template(body)
		v, err := strconv.Atoi(res)
		assert.NoError(t, err)
		assert.True(t, v > 0)

	})

	t.Run("rnd_int", func(t *testing.T) {
		body := `{{rnd_int}}`
		res := x.Template(body)

		v, err := strconv.Atoi(res)
		assert.NoError(t, err)
		assert.True(t, v > 0)
	})

	t.Run("unix", func(t *testing.T) {
		body := `{{unix}}`
		res := x.Template(body)

		v, err := strconv.Atoi(res)
		assert.NoError(t, err)
		assert.True(t, v > 0)
	})

	t.Run("md5", func(t *testing.T) {
		const expect = `869bc90a958424fd95dcc0d57d14be6f`
		body := `{{ md5 "a=b&c=d"}}`
		res := x.Template(body)
		assert.Equal(t, expect, res)
	})

	t.Run("query", func(t *testing.T) {
		const expect = `a=b&c=d`
		body := `{{ query "c" "d" "a" "b"}}`
		res := x.Template(body)
		assert.Equal(t, expect, res)
	})

	t.Run("combination", func(t *testing.T) {
		q := url.Values{"a": []string{"b"}}
		target := fmt.Sprintf("%x", md5.Sum([]byte(q.Encode())))

		body := `{{ md5 (query "a" "b") }}`
		res := x.Template(body)
		assert.Equal(t, target, res)
	})

	// concatenate result of function with variable + put output in other function
	t.Run("advance concatenate + combination", func(t *testing.T) {
		const v = "GOGO"
		q := url.Values{"a": []string{"b"}}
		target := fmt.Sprintf("%x", md5.Sum([]byte(q.Encode()+v)))

		body := `{{ (printf "%s%s" (query "a" "b") .HELLO ) | md5 }}`
		x := New(map[string]string{"HELLO": v}, nil)
		res := x.Template(body)
		assert.Equal(t, target, res)
	})

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
				if err == nil {
					assert.NotEmpty(t, res)
				}
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
			assert.NotEmpty(t, res)
			assert.True(t, string(res) == "10.00" || string(res) == "0.00")
		}
	})
}
