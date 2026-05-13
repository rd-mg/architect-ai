package catalog

import "github.com/rd-mg/architect-ai/internal/model"

type Skill struct {
	ID       model.SkillID
	Name     string
	Category string
	Priority string
}

func MVPSkills() []Skill {
	all := model.AllSkills()
	skills := make([]Skill, len(all))
	for i, sr := range all {
		skills[i] = Skill{
			ID:       sr.ID,
			Name:     sr.Name,
			Category: sr.Category,
			Priority: sr.Priority,
		}
	}
	return skills
}
