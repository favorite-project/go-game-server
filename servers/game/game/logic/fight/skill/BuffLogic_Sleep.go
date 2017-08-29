package skill

/*
睡眠
*/

type BuffLogic_Sleep struct {
	BuffLogic_Base
}

func (bl *BuffLogic_Sleep) CanAttack() bool {
	return false
}

func (bl *BuffLogic_Sleep) CanUseSkill() bool {
	return false
}

func (bl *BuffLogic_Sleep) CanBeInterrupt() bool {
	return true
}
