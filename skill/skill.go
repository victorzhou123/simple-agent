package skill

type Skill interface {
	GetName() string
	GetDescription() string
	GetContent() string
	IsAvailable() bool
	UpdateAvailability()
}

type skill struct {
	name        string
	available   *bool
	path        string
	description string
	content     string
}

func newSkill(name, path, description, content string) Skill {
	return &skill{
		name:        name,
		path:        path,
		description: description,
		content:     content,
	}
}

func (s *skill) IsAvailable() bool {
	if s.available == nil {
		return false
	}
	return *s.available
}

func (s *skill) GetName() string {
	return s.name
}

func (s *skill) GetPath() string {
	return s.path
}

func (s *skill) GetDescription() string {
	return s.description
}

func (s *skill) GetContent() string {
	return s.content
}

func (s *skill) UpdateAvailability() {
	var available bool

	if s.name == "" {
		s.available = &available
		return
	}

	if s.path == "" {
		s.available = &available
		return
	}

	if s.description == "" {
		s.available = &available
		return
	}

	if s.content == "" {
		s.available = &available
		return
	}

	available = true
	s.available = &available
}
