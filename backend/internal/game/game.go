package game

import (
	"errors"
	"fmt"
	"math"
)

const HeroMaxHP int64 = 200

type State struct {
	HeroHP         int64   `json:"heroHP"`
	BossHP         int64   `json:"bossHP"`
	MonsterHP      []int64 `json:"monsterHP"`
	TreasureOpened int64   `json:"treasureOpened"`
}

type Spec struct {
	Monsters   int64  `json:"monsters"`
	Difficulty string `json:"difficulty"`
	HeroClass  string `json:"heroClass"`
	AttackSeq  int64  `json:"attackSeq"`
	Game       State  `json:"game"`
}

type Rules struct {
	MonsterMaxHP   int64
	BossMaxHP      int64
	HeroDamage     int64
	MonsterCounter int64
	BossCounter    int64
}

var (
	ErrHeroDefeated      = errors.New("hero is defeated; create a new dungeon")
	ErrCombatComplete    = errors.New("boss is defeated; combat is complete")
	ErrBossLocked        = errors.New("boss is locked until all monsters are defeated")
	ErrTargetDefeated    = errors.New("target is already defeated")
	ErrInvalidTarget     = errors.New("target must be this dungeon's boss or an exact monster name")
	ErrTreasureLocked    = errors.New("treasure is locked until the boss is defeated")
	ErrTreasureOpened    = errors.New("treasure is already open")
	ErrInvalidAction     = errors.New("only open-treasure is supported")
	ErrSequenceExhausted = errors.New("turn sequence exhausted; create a new dungeon")
)

func RulesForDifficulty(difficulty string) (Rules, bool) {
	switch difficulty {
	case "easy":
		return Rules{30, 200, 25, 1, 3}, true
	case "normal":
		return Rules{50, 400, 35, 2, 5}, true
	case "hard":
		return Rules{80, 800, 45, 3, 8}, true
	default:
		return Rules{}, false
	}
}

func NewSpec(monsters int64, difficulty, heroClass string) (Spec, error) {
	rules, ok := RulesForDifficulty(difficulty)
	if !ok {
		return Spec{}, errors.New("difficulty must be easy, normal, or hard")
	}
	if monsters < 1 || monsters > 10 {
		return Spec{}, errors.New("monsters must be between 1 and 10")
	}
	if heroClass == "" {
		heroClass = "warrior"
	}
	if heroClass != "warrior" {
		return Spec{}, errors.New("heroClass must be warrior")
	}
	hp := make([]int64, monsters)
	for i := range hp {
		hp[i] = rules.MonsterMaxHP
	}
	return Spec{
		Monsters: monsters, Difficulty: difficulty, HeroClass: heroClass,
		Game: State{HeroHP: HeroMaxHP, BossHP: rules.BossMaxHP, MonsterHP: hp},
	}, nil
}

func (s Spec) Validate() error {
	rules, ok := RulesForDifficulty(s.Difficulty)
	if !ok || s.HeroClass != "warrior" || s.Monsters < 1 || s.Monsters > 10 {
		return errors.New("invalid dungeon configuration")
	}
	if s.AttackSeq < 0 || int64(len(s.Game.MonsterHP)) != s.Monsters {
		return errors.New("invalid dungeon turn sequence or monster HP array")
	}
	if s.Game.HeroHP < 0 || s.Game.HeroHP > HeroMaxHP ||
		s.Game.BossHP < 0 || s.Game.BossHP > rules.BossMaxHP ||
		s.Game.TreasureOpened < 0 || s.Game.TreasureOpened > 1 {
		return errors.New("invalid dungeon game state")
	}
	for _, hp := range s.Game.MonsterHP {
		if hp < 0 || hp > rules.MonsterMaxHP {
			return errors.New("invalid monster HP")
		}
	}
	return nil
}

func (s Spec) Attack(name, target string) (Spec, error) {
	if err := s.Validate(); err != nil {
		return s, err
	}
	if s.Game.HeroHP == 0 {
		return s, ErrHeroDefeated
	}
	if s.Game.BossHP == 0 {
		return s, ErrCombatComplete
	}
	if s.AttackSeq == math.MaxInt64 {
		return s, ErrSequenceExhausted
	}
	rules, _ := RulesForDifficulty(s.Difficulty)
	next := s.clone()
	var counter int64
	if target == name+"-boss" {
		for _, hp := range s.Game.MonsterHP {
			if hp > 0 {
				return s, ErrBossLocked
			}
		}
		next.Game.BossHP = max(0, s.Game.BossHP-rules.HeroDamage)
		if next.Game.BossHP > 0 {
			counter = rules.BossCounter * PhaseMultiplier(next.Game.BossHP, rules.BossMaxHP) / 10
		}
	} else {
		index := -1
		for i := range s.Game.MonsterHP {
			if target == fmt.Sprintf("%s-monster-%d", name, i) {
				index = i
				break
			}
		}
		if index < 0 {
			return s, ErrInvalidTarget
		}
		if s.Game.MonsterHP[index] == 0 {
			return s, ErrTargetDefeated
		}
		next.Game.MonsterHP[index] = max(0, s.Game.MonsterHP[index]-rules.HeroDamage)
		for _, hp := range next.Game.MonsterHP {
			if hp > 0 {
				counter += rules.MonsterCounter
			}
		}
	}
	// Surviving monsters counter as one chain. Apply the warrior reduction to
	// the total, then round down once rather than rounding each small hit.
	next.Game.HeroHP = max(0, s.Game.HeroHP-counter*3/4)
	next.AttackSeq++
	return next, nil
}

func (s Spec) Action(action string) (Spec, error) {
	if err := s.Validate(); err != nil {
		return s, err
	}
	if action != "open-treasure" {
		return s, ErrInvalidAction
	}
	if s.Game.HeroHP == 0 {
		return s, ErrHeroDefeated
	}
	if s.Game.BossHP > 0 {
		return s, ErrTreasureLocked
	}
	if s.Game.TreasureOpened == 1 {
		return s, ErrTreasureOpened
	}
	if s.AttackSeq == math.MaxInt64 {
		return s, ErrSequenceExhausted
	}
	next := s.clone()
	next.Game.TreasureOpened = 1
	next.AttackSeq++
	return next, nil
}

func PhaseMultiplier(hp, maxHP int64) int64 {
	if hp*100 > maxHP*50 {
		return 10
	}
	if hp*100 > maxHP*25 {
		return 13
	}
	return 16
}

func (s Spec) clone() Spec {
	s.Game.MonsterHP = append([]int64(nil), s.Game.MonsterHP...)
	return s
}
