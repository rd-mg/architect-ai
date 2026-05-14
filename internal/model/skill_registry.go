package model

type SkillRegistration struct {
	ID       SkillID
	Name     string
	Category string
	Priority string
	Preset   string // "minimal", "ecosystem", "full" — minimum preset that includes this skill
}

var skillRegistrations []SkillRegistration
var skillByID = make(map[SkillID]*SkillRegistration)

func RegisterSkill(sr SkillRegistration) {
	if _, exists := skillByID[sr.ID]; exists {
		panic("duplicate skill registration: " + string(sr.ID))
	}
	skillRegistrations = append(skillRegistrations, sr)
	skillByID[sr.ID] = &skillRegistrations[len(skillRegistrations)-1]
}

func GetSkill(id SkillID) *SkillRegistration {
	return skillByID[id]
}

func AllSkills() []SkillRegistration {
	result := make([]SkillRegistration, len(skillRegistrations))
	copy(result, skillRegistrations)
	return result
}

func SkillsForCategory(category string) []SkillRegistration {
	var result []SkillRegistration
	for _, sr := range skillRegistrations {
		if sr.Category == category {
			result = append(result, sr)
		}
	}
	return result
}

var presetTiers = map[string]int{"minimal": 0, "ecosystem": 1, "full": 2}

func presetTargetTier(preset PresetID) int {
	switch preset {
	case PresetMinimal:
		return 0
	case PresetEcosystemOnly:
		return 1
	case PresetFullGentleman:
		return 2
	case PresetCustom:
		return -1
	default:
		return 2
	}
}

func SkillsForPreset(preset PresetID) []SkillID {
	targetTier := presetTargetTier(preset)
	if targetTier < 0 {
		return nil
	}
	var result []SkillID
	for _, sr := range skillRegistrations {
		if skillTier, ok := presetTiers[sr.Preset]; ok && skillTier <= targetTier {
			result = append(result, sr.ID)
		}
	}
	return result
}

func AllSkillIDs() []SkillID {
	result := make([]SkillID, len(skillRegistrations))
	for i, sr := range skillRegistrations {
		result[i] = sr.ID
	}
	return result
}

func init() {
	RegisterSkill(SkillRegistration{ID: SkillSDDInit, Name: "sdd-init", Category: "sdd", Priority: "p0", Preset: "minimal"})
	RegisterSkill(SkillRegistration{ID: SkillSDDExplore, Name: "sdd-explore", Category: "sdd", Priority: "p0", Preset: "minimal"})
	RegisterSkill(SkillRegistration{ID: SkillSDDPropose, Name: "sdd-propose", Category: "sdd", Priority: "p0", Preset: "minimal"})
	RegisterSkill(SkillRegistration{ID: SkillSDDSpec, Name: "sdd-spec", Category: "sdd", Priority: "p0", Preset: "minimal"})
	RegisterSkill(SkillRegistration{ID: SkillSDDDesign, Name: "sdd-design", Category: "sdd", Priority: "p0", Preset: "minimal"})
	RegisterSkill(SkillRegistration{ID: SkillSDDTasks, Name: "sdd-tasks", Category: "sdd", Priority: "p0", Preset: "minimal"})
	RegisterSkill(SkillRegistration{ID: SkillSDDApply, Name: "sdd-apply", Category: "sdd", Priority: "p0", Preset: "minimal"})
	RegisterSkill(SkillRegistration{ID: SkillSDDVerify, Name: "sdd-verify", Category: "sdd", Priority: "p0", Preset: "minimal"})
	RegisterSkill(SkillRegistration{ID: SkillSDDArchive, Name: "sdd-archive", Category: "sdd", Priority: "p0", Preset: "minimal"})
	RegisterSkill(SkillRegistration{ID: SkillSDDOnboard, Name: "sdd-onboard", Category: "sdd", Priority: "p0", Preset: "minimal"})
	RegisterSkill(SkillRegistration{ID: SkillJudgmentDay, Name: "judgment-day", Category: "sdd", Priority: "p0", Preset: "minimal"})
	RegisterSkill(SkillRegistration{ID: SkillSolver, Name: "solver", Category: "specialist", Priority: "p0", Preset: "full"})
	RegisterSkill(SkillRegistration{ID: SkillIdeator, Name: "ideator", Category: "specialist", Priority: "p0", Preset: "full"})
	RegisterSkill(SkillRegistration{ID: SkillResearcher, Name: "researcher", Category: "specialist", Priority: "p0", Preset: "full"})
	RegisterSkill(SkillRegistration{ID: SkillGeneralist, Name: "generalist", Category: "specialist", Priority: "p0", Preset: "full"})
	RegisterSkill(SkillRegistration{ID: SkillArchitectureGuardrails, Name: "architecture-guardrails", Category: "specialist", Priority: "p0", Preset: "full"})
	RegisterSkill(SkillRegistration{ID: SkillGoTesting, Name: "go-testing", Category: "testing", Priority: "p0", Preset: "ecosystem"})
	RegisterSkill(SkillRegistration{ID: SkillCreator, Name: "skill-creator", Category: "workflow", Priority: "p0", Preset: "ecosystem"})
	RegisterSkill(SkillRegistration{ID: SkillBranchPR, Name: "branch-pr", Category: "workflow", Priority: "p0", Preset: "ecosystem"})
	RegisterSkill(SkillRegistration{ID: SkillIssueCreation, Name: "issue-creation", Category: "workflow", Priority: "p0", Preset: "ecosystem"})
	RegisterSkill(SkillRegistration{ID: SkillSkillRegistry, Name: "skill-registry", Category: "workflow", Priority: "p0", Preset: "ecosystem"})
}
