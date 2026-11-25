package flow

var _ DataStore = (*MemDataStore)(nil)

type MemDataStore struct {
	data map[string]any
}

func (m *MemDataStore) Get(key string) (any, bool) {
	if m.Has(key) {
		return m.data[key], true
	}

	return nil, false
}

func (m *MemDataStore) Has(key string) bool {
	_, ok := m.data[key]
	return ok
}

func (m *MemDataStore) Keys() []string {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}

	return keys
}

func (m *MemDataStore) Set(key string, value any) {
	m.data[key] = value
}

func (m *MemDataStore) Delete(key string) {
	delete(m.data, key)
}

func NewMemDataStore() *MemDataStore {
	return &MemDataStore{
		data: make(map[string]any),
	}
}
