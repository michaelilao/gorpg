package game

type Monster struct {
	Character
}

func NewRat(p Pos) *Monster {
	monster := &Monster{}
	monster.Pos = p
	monster.Rune = 'R'
	monster.Name = "Rat"
	monster.Hitpoints = 50
	monster.Strength = 5
	monster.Speed = 1.5
	monster.ActionPoints = 0.0
	monster.SightRange = 10

	return monster
}

func NewSpider(p Pos) *Monster {
	monster := &Monster{}
	monster.Pos = p
	monster.Rune = 'S'
	monster.Name = "Spider"
	monster.Hitpoints = 100
	monster.Strength = 5
	monster.Speed = 1.0
	monster.ActionPoints = 0.0
	monster.SightRange = 10
	return monster
}
func (m *Monster) Pass() {
	m.ActionPoints -= m.Speed
}

func (m *Monster) Update(level *Level) {
	m.ActionPoints += m.Speed
	playerPos := level.Player.Pos

	apInt := int(m.ActionPoints)
	positions := level.astar(m.Pos, playerPos)
	moveIndex := 1

	if len(positions) == 0 {
		m.Pass()
		return
	}
	for i := 0; i < apInt; i++ {
		if moveIndex < len(positions) {
			m.Move(positions[moveIndex], level)
			moveIndex++
			m.ActionPoints--
		}
	}
}

func (m *Monster) Move(to Pos, level *Level) {
	_, exists := level.Monsters[to]
	if !exists && to != level.Player.Pos {
		delete(level.Monsters, m.Pos)
		level.Monsters[to] = m
		m.Pos = to
		return
	}
	if to == level.Player.Pos {
		level.Attack(&m.Character, &level.Player.Character)
		if m.Hitpoints <= 0 {
			delete(level.Monsters, m.Pos)
		}
		if level.Player.Hitpoints <= 0 {
			level.AddEvent("You have died")
		}
	}
}
