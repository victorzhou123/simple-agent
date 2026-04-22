package loadskill

import (
	"context"
	"fmt"

	"simple-agent/skill"
	"simple-agent/tools/base"
)

func New(cfg base.ToolConfig, skillMgr skill.SkillManager) base.Tool {
	return &loadSkillTool{
		BaseTool: base.NewBaseTool(cfg.Name, cfg.Description, cfg.InputSchema),
		skillMgr: skillMgr,
	}
}

type loadSkillTool struct {
	base.BaseTool
	skillMgr skill.SkillManager
}

func (t *loadSkillTool) Call(ctx context.Context, args map[string]any) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("load_skill: name must be a non-empty string")
	}

	sk := t.skillMgr.GetSkill(name)
	if sk == nil {
		return "", fmt.Errorf("load_skill: skill '%s' not found", name)
	}

	if !sk.IsAvailable() {
		return "", fmt.Errorf("load_skill: skill '%s' is not available", name)
	}

	return fmt.Sprintf("📚 Skill: %s\n\n%s", name, sk.GetContent()), nil
}
