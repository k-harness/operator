package variables

type Store struct {
	store map[string]string
}

// New Store merge status variables with secrets vars
// status variables has greater priority
func New(status, protected map[string]string) *Store {
	store := make(map[string]string)
	for k, v := range protected {
		store[k] = v
	}

	// over
	for k, v := range status {
		store[k] = v
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
