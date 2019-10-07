package storage

type Entity struct {
	Sid   string
	Email string
}

type Storage struct {
	description string
	m           map[string]*Entity
}

func New(description string) *Storage {
	return &Storage{
		description: description,
		m:           map[string]*Entity{},
	}
}
func (s *Storage) InitUserData() {
	data := []*Entity{
		&Entity{Sid: "1dd8e7c2-0cb2-4bc0-bd3b-5f6ca2db91cd", Email: "Brady"},
		&Entity{Sid: "dc71f897-1286-4251-9751-a7c7c3213335", Email: "Marie"},
		&Entity{Sid: "3d951fe1-5486-4999-878f-ae98ccd8afe3", Email: "Donny"},
		&Entity{Sid: "ab17ca3b-c844-4b54-9626-75a577219504", Email: "Marsha"},
		&Entity{Sid: "0ec0fa7-b6e2-4ee8-b06c-8991e2446f07", Email: "Jan"},
	}

	for _, e := range data {
		s.m[e.Sid] = e
	}
}

func (s *Storage) GetEntities() []*Entity {
	res := []*Entity{}
	for _, e := range s.m {
		res = append(res, e)
	}

	return res
}
