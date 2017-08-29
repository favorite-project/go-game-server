package skill

/*
禁法
*/
type BuffLogic_SkillDisable struct {
	BuffLogic_Base
}

func (bl *BuffLogic_SkillDisable) CanUseSkill() bool {
	return false
}
