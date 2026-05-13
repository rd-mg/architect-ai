package skills

import "github.com/rd-mg/architect-ai/internal/model"

func SkillsForPreset(preset model.PresetID) []model.SkillID {
	return model.SkillsForPreset(preset)
}

func AllSkillIDs() []model.SkillID {
	return model.AllSkillIDs()
}
