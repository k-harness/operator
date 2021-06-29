package variables

import (
	"bytes"
	"text/template"
)

type Store struct {
	store map[string]string
}

// New Store merge status variables with secrets vars
// status variables has greater priority
func New(in ...map[string]string) *Store {
	store := make(map[string]string)

	buf := bytes.NewBuffer(nil)

	// Note: super slow flow, i req. use sync.Map as this proc every work
	for _, maps := range in {
		for k, v := range maps {
			buf.Reset()

			t, err := template.New("x").
				Funcs(TemplateFunctions).
				Parse(v)
			if err == nil {
				if err = t.Execute(buf, store); err == nil {
					v = buf.String()
				}
			}

			store[k] = v
		}
	}

	return &Store{store: store}
}

func (s *Store) Update(k, v string) {
	s.store[k] = v
}

func (s *Store) Get(key string) (string, bool) {
	v, ok := s.store[key]
	return v, ok
}
