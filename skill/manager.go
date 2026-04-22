package skill

import (
	"path/filepath"
	"regexp"
	"strings"

	"simple-agent/utils"
)

type skillManager struct {
	documents map[string]Skill
}

type SkillManager interface {
	DescribeAvailableSkills() string
	GetSkill(name string) Skill
}

func NewSkillManager() SkillManager {
	// project skills
	skillFiles := searchSkillMDFilesPath(utils.SkillDir)
	documents := make(map[string]Skill)
	for _, file := range skillFiles {
		name, skill := loadSingleSkill(file)
		if skill != nil {
			documents[name] = skill
		}
	}

	return &skillManager{
		documents: documents,
	}
}

func (s *skillManager) DescribeAvailableSkills() string {
	var describe strings.Builder

	for name, skill := range s.documents {
		if skill.IsAvailable() {
			describe.WriteString(name + ": " + skill.GetDescription() + "\n")
		}
	}

	return describe.String() + "\n"
}

func (s *skillManager) GetSkill(name string) Skill {
	return s.documents[name]
}

func loadSingleSkill(path string) (string, Skill) {
	doc, err := utils.LoadFile(path)
	if err != nil {
		return "", nil
	}

	skill, err := parseSkillDocument(doc, path)
	if err != nil {
		return "", nil
	}
	return skill.GetName(), skill
}

func parseFrontmatter(text string) (map[string]string, string) {
	re := regexp.MustCompile(`(?s)^---\n(.*?)\n---\n(.*)`)
	matches := re.FindStringSubmatch(text)

	if matches == nil {
		return map[string]string{}, text
	}

	meta := make(map[string]string)
	frontmatter := strings.TrimSpace(matches[1])
	body := matches[2]

	for _, line := range strings.Split(frontmatter, "\n") {
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			meta[key] = value
		}
	}

	return meta, body
}

func parseSkillDocument(doc, path string) (Skill, error) {
	meta, body := parseFrontmatter(doc)

	name := meta["name"]
	description := meta["description"]

	sk := newSkill(name, path, description, body)

	sk.UpdateAvailability()

	return sk, nil
}

func searchSkillMDFilesPath(dirPath string) []string {
	paths, err := utils.GetAllFilePaths(dirPath, -1)
	if err != nil {
		return nil
	}
	var skillFiles []string
	for _, path := range paths {
		if filepath.Base(path) == "SKILL.md" || filepath.Base(path) == "skill.md" {
			skillFiles = append(skillFiles, path)
		}
	}
	return skillFiles
}
