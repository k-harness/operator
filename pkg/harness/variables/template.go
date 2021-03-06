package variables

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/url"
	"text/template"
	"time"

	"log"

	"github.com/google/uuid"
	"github.com/k-harness/operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/rand"
)

var TemplateFunctions = template.FuncMap{
	"uuid": func() string {
		return uuid.New().String()
	},
	"unix": func() int {
		return int(time.Now().Unix())
	},
	"rnd_int": func() int {
		return rand.Int()
	},
	"rnd_str": func(len int) string {
		return rand.String(len)
	},
	"range_int": func(min, max int) int {
		return rand.IntnRange(min, max)
	},
	"md5": func(in string) string {
		return fmt.Sprintf("%x", md5.Sum([]byte(in)))
	},
	"query": func(args ...string) string {
		vals := make(url.Values)
		for i := 0; i < len(args); i += 2 {
			vals.Add(args[i], args[i+1])
		}

		return vals.Encode()
	},
}

func (s *Store) TemplateBytesOrReturnWithout(in []byte) []byte {
	r, err := s.TemplateBytes(in)
	if err != nil {
		log.Printf("[ERR]: variable store was not able template body %q error %s", string(in), err)
		return in
	}

	return r
}

func (s *Store) Template(in string) string {
	return string(s.TemplateBytesOrReturnWithout([]byte(in)))
}

func (s *Store) TemplateBytes(in []byte) ([]byte, error) {
	t, err := template.New("x").
		Funcs(TemplateFunctions).
		Parse(string(in))
	if err != nil {
		return nil, fmt.Errorf("template parse error: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	if err = t.Execute(buf, s.store); err != nil {
		return nil, fmt.Errorf("store template executor: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *Store) TemplateMapOrReturnWhatPossible(in map[string]string) map[string]string {
	res := make(map[string]string)

	for k, v := range in {
		key, err := s.TemplateBytes([]byte(k))
		if err != nil {
			log.Printf("[ERR]: template map key %q error: %s", k, err)
			key = []byte(err.Error())
		}

		val, err := s.TemplateBytes([]byte(v))
		if err != nil {
			log.Printf("[ERR]: template map val %q  of key %q error: %s", v, k, err)
			val = []byte(err.Error())
		}

		res[string(key)] = string(val)
	}

	return res
}

func (s *Store) RequestTranslate(in *v1alpha1.Request) {
	in.Header = s.TemplateMapOrReturnWhatPossible(in.Header)

	if len(in.Body.KV) > 0 {
		in.Body.KV = s.TemplateMapOrReturnWhatPossible(in.Body.KV)
	}

	if len(in.Body.Byte) > 0 {
		in.Body.Byte = s.TemplateBytesOrReturnWithout(in.Body.Byte)
	}

	if len(in.Body.Row) > 0 {
		in.Body.Row = s.Template(in.Body.Row)
	}
}
