package game

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestNewSpec(t *testing.T) {
	for _, tt := range []struct {
		difficulty string
		monsterHP  int64
		bossHP     int64
	}{
		{"easy", 30, 200}, {"normal", 50, 400}, {"hard", 80, 800},
	} {
		t.Run(tt.difficulty, func(t *testing.T) {
			s, err := NewSpec(3, tt.difficulty, "")
			if err != nil {
				t.Fatal(err)
			}
			if s.HeroClass != "warrior" || s.AttackSeq != 0 || s.Game.HeroHP != 200 ||
				s.Game.BossHP != tt.bossHP || s.Game.TreasureOpened != 0 ||
				!reflect.DeepEqual(s.Game.MonsterHP, []int64{tt.monsterHP, tt.monsterHP, tt.monsterHP}) {
				t.Fatalf("unexpected initial state: %+v", s)
			}
			if err := s.Validate(); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, tt := range []struct {
		monsters   int64
		difficulty string
		heroClass  string
	}{
		{0, "easy", "warrior"}, {11, "easy", "warrior"},
		{1, "invalid", "warrior"}, {1, "easy", "mage"}, {1, "easy", "rogue"},
	} {
		if _, err := NewSpec(tt.monsters, tt.difficulty, tt.heroClass); err == nil {
			t.Errorf("accepted invalid configuration %+v", tt)
		}
	}
}

func TestAttackDamageAndNoMutation(t *testing.T) {
	for _, tt := range []struct {
		difficulty string
		hp         int64
		heroHP     int64
	}{
		{"easy", 5, 200}, {"normal", 15, 199}, {"hard", 35, 198},
	} {
		t.Run(tt.difficulty, func(t *testing.T) {
			s, _ := NewSpec(1, tt.difficulty, "warrior")
			before := s.clone()
			next, err := s.Attack("boss-room", "boss-room-monster-0")
			if err != nil {
				t.Fatal(err)
			}
			if next.Game.MonsterHP[0] != tt.hp || next.Game.HeroHP != tt.heroHP || next.AttackSeq != 1 {
				t.Fatalf("unexpected result %+v", next)
			}
			if !reflect.DeepEqual(s, before) {
				t.Fatalf("attack mutated original: %+v", s)
			}
			killed, err := next.Attack("boss-room", "boss-room-monster-0")
			if err != nil {
				t.Fatal(err)
			}
			if killed.Game.MonsterHP[0] != 0 || killed.Game.HeroHP != next.Game.HeroHP || killed.AttackSeq != 2 {
				t.Fatalf("killing hit must clamp HP and have no counter: %+v", killed)
			}
		})
	}
}

func TestExactTargetNames(t *testing.T) {
	s, _ := NewSpec(2, "easy", "warrior")
	for _, target := range []string{
		"", "boss", "other-boss", "the-boss-monster-00", "the-boss-monster--1",
		"the-boss-monster-2", "other-monster-0", "the-boss-monster-0-extra",
		"open-treasure",
	} {
		if _, err := s.Attack("the-boss", target); !errors.Is(err, ErrInvalidTarget) {
			t.Errorf("target %q: got %v, want invalid target", target, err)
		}
	}
	if _, err := s.Attack("the-boss", "the-boss-boss"); !errors.Is(err, ErrBossLocked) {
		t.Fatalf("boss target was not locked: %v", err)
	}
}

func TestBossCounterPhases(t *testing.T) {
	for _, tt := range []struct {
		hpAfter    int64
		multiplier int64
		heroHP     int64
	}{
		{401, 10, 194}, {400, 13, 193}, {201, 13, 193}, {200, 16, 191}, {1, 16, 191},
	} {
		s, _ := NewSpec(1, "hard", "warrior")
		s.Game.MonsterHP[0] = 0
		s.Game.BossHP = tt.hpAfter + 45
		next, err := s.Attack("room", "room-boss")
		if err != nil {
			t.Fatal(err)
		}
		if next.Game.BossHP != tt.hpAfter || next.Game.HeroHP != tt.heroHP ||
			PhaseMultiplier(tt.hpAfter, 800) != tt.multiplier {
			t.Errorf("boss HP %d: unexpected result %+v", tt.hpAfter, next)
		}
	}
}

func TestHealthFloorsAndTerminalStates(t *testing.T) {
	s, _ := NewSpec(1, "hard", "warrior")
	s.Game.HeroHP = 1
	next, err := s.Attack("room", "room-monster-0")
	if err != nil || next.Game.HeroHP != 0 {
		t.Fatalf("hero HP must clamp to zero: %+v, %v", next, err)
	}
	if _, err := next.Attack("room", "room-monster-0"); !errors.Is(err, ErrHeroDefeated) {
		t.Fatalf("defeated hero accepted combat: %v", err)
	}
	if _, err := next.Action("open-treasure"); !errors.Is(err, ErrHeroDefeated) {
		t.Fatalf("defeated hero accepted action: %v", err)
	}
	s.Game.HeroHP = 200
	s.Game.MonsterHP[0] = 0
	if _, err := s.Attack("room", "room-monster-0"); !errors.Is(err, ErrTargetDefeated) {
		t.Fatalf("dead target accepted combat: %v", err)
	}
	s.Game.BossHP = 1
	next, err = s.Attack("room", "room-boss")
	if err != nil || next.Game.BossHP != 0 || next.Game.HeroHP != 200 {
		t.Fatalf("boss killing hit incorrect: %+v, %v", next, err)
	}
	if _, err := next.Attack("room", "room-boss"); !errors.Is(err, ErrCombatComplete) {
		t.Fatalf("completed dungeon accepted combat: %v", err)
	}
}

func TestActionsAndSharedSequence(t *testing.T) {
	s, _ := NewSpec(1, "easy", "warrior")
	if _, err := s.Action("enter-room-2"); !errors.Is(err, ErrInvalidAction) {
		t.Fatalf("unsupported action accepted: %v", err)
	}
	if _, err := s.Action("open-treasure"); !errors.Is(err, ErrTreasureLocked) {
		t.Fatalf("treasure opened before boss kill: %v", err)
	}
	s.Game.MonsterHP[0] = 0
	s.Game.BossHP = 1
	s, _ = s.Attack("room", "room-boss")
	next, err := s.Action("open-treasure")
	if err != nil || next.Game.TreasureOpened != 1 || next.AttackSeq != 2 {
		t.Fatalf("unexpected action result %+v, %v", next, err)
	}
	if s.Game.TreasureOpened != 0 {
		t.Fatal("action mutated original")
	}
	if _, err := next.Action("open-treasure"); !errors.Is(err, ErrTreasureOpened) {
		t.Fatalf("repeated treasure action accepted: %v", err)
	}
	s.AttackSeq = math.MaxInt64
	if _, err := s.Action("open-treasure"); !errors.Is(err, ErrSequenceExhausted) {
		t.Fatalf("action sequence overflow accepted: %v", err)
	}
	s.Game.BossHP = 1
	if _, err := s.Attack("room", "room-boss"); !errors.Is(err, ErrSequenceExhausted) {
		t.Fatalf("combat sequence overflow accepted: %v", err)
	}
}

func TestInvalidStoredState(t *testing.T) {
	for _, mutate := range []func(*Spec){
		func(s *Spec) { s.Monsters = 0 },
		func(s *Spec) { s.Difficulty = "invalid" },
		func(s *Spec) { s.HeroClass = "mage" },
		func(s *Spec) { s.AttackSeq = -1 },
		func(s *Spec) { s.Game.MonsterHP = nil },
		func(s *Spec) { s.Game.MonsterHP[0] = -1 },
		func(s *Spec) { s.Game.MonsterHP[0] = 1000 },
		func(s *Spec) { s.Game.HeroHP = -1 },
		func(s *Spec) { s.Game.HeroHP = 201 },
		func(s *Spec) { s.Game.BossHP = -1 },
		func(s *Spec) { s.Game.BossHP = 801 },
		func(s *Spec) { s.Game.TreasureOpened = 2 },
	} {
		s, _ := NewSpec(1, "hard", "warrior")
		mutate(&s)
		if err := s.Validate(); err == nil {
			t.Errorf("accepted corrupt state %+v", s)
		}
		if _, err := s.Attack("room", "room-monster-0"); err == nil {
			t.Errorf("combat accepted corrupt state %+v", s)
		}
		if _, err := s.Action("open-treasure"); err == nil {
			t.Errorf("action accepted corrupt state %+v", s)
		}
	}
}

func TestFullRunAllDifficulties(t *testing.T) {
	for _, difficulty := range []string{"easy", "normal", "hard"} {
		t.Run(difficulty, func(t *testing.T) {
			s, _ := NewSpec(2, difficulty, "warrior")
			for i := range s.Game.MonsterHP {
				for s.Game.MonsterHP[i] > 0 {
					var err error
					s, err = s.Attack("room", fmt.Sprintf("room-monster-%d", i))
					if err != nil {
						t.Fatal(err)
					}

				}
			}
			for s.Game.BossHP > 0 {
				var err error
				s, err = s.Attack("room", "room-boss")
				if err != nil {
					t.Fatal(err)
				}
			}
			if s.Game.HeroHP <= 0 {
				t.Fatal("two-monster demo must remain winnable")
			}
			s, err := s.Action("open-treasure")
			if err != nil || s.Game.TreasureOpened != 1 {
				t.Fatalf("could not finish run: %+v, %v", s, err)
			}
		})
	}
}

func TestHardTenMonstersCanDefeatHero(t *testing.T) {
	s, _ := NewSpec(10, "hard", "warrior")
	first, err := s.Attack("room", "room-monster-0")
	if err != nil || first.Game.HeroHP != 178 {
		t.Fatalf("all ten monsters must counter, reduced once: %+v, %v", first, err)
	}
	s = first
	for i := range s.Game.MonsterHP {
		for s.Game.MonsterHP[i] > 0 && s.Game.HeroHP > 0 {
			var err error
			s, err = s.Attack("room", fmt.Sprintf("room-monster-%d", i))
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if s.Game.HeroHP != 0 || s.Game.BossHP != 800 || s.AttackSeq <= 1 {
		t.Fatalf("expected a repeatable legal defeat: %+v", s)
	}
	if _, err := s.Attack("room", "room-monster-9"); !errors.Is(err, ErrHeroDefeated) {
		t.Fatalf("combat after legal defeat accepted: %v", err)
	}
	if _, err := s.Action("open-treasure"); !errors.Is(err, ErrHeroDefeated) {
		t.Fatalf("treasure after legal defeat accepted: %v", err)
	}
}
