package skill

/*
无敌
*/
type BuffLogic_Invincible struct {
	BuffLogic_Base
}

func (bl *BuffLogic_Invincible) CanBeAttacked() bool {
	return false
}
