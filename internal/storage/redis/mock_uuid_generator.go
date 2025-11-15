package redis

type MockUUIDGenerator struct {
	uuid string
}

func (m *MockUUIDGenerator) New() string {
	return m.uuid
}
