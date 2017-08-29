package skill

/*
眩晕
*/
type BuffLogic_Dizzy struct {
	BuffLogic_Base
}

func (bl *BuffLogic_Dizzy) CanAttack() bool {
	return false
}

func (bl *BuffLogic_Dizzy) CanUseSkill() bool {
	return false
}
