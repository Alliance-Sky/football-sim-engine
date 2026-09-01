package engine

import (
	"math/rand"
	"testing"
)

func createTestTeam(name, formation, kitColor string, baseRating float64) (*Team, error) {
	slots := FormationSlots[formation]
	var players []*Player
	i := 1
	for pos, count := range slots {
		for _ = range count {
			p, err := NewPlayer(
				name+"_"+string(pos),
				name+" "+string(pos),
				pos,
				FootBoth,
				baseRating,
				24,
				100.0,
				nil,
				90.0,
			)
			if err != nil {
				return nil, err
			}
			p.AssignedPosition = pos
			players = append(players, p)
			i++
		}
	}
	return NewTeam(name+"_id", name, formation, kitColor, players)
}

func TestOOPPenalty(t *testing.T) {
	// Natural position -> 0 penalty
	p, err := NewPlayer("p1", "Player 1", PosLB, FootBoth, 80.0, 25, 100.0, nil,
					90.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.EffectiveRating() != 80.0 {
		t.Errorf("expected 80.0, got %.1f", p.EffectiveRating())
	}

	// Outfielder in GK -> 50% penalty
	pGK, _ := NewPlayer("p2", "Player 2", PosST, FootBoth, 80.0, 25, 100.0, new(PosGK), 90.0)
	if pGK.EffectiveRating() != 40.0 {
		t.Errorf("expected 40.0 for ST playing GK, got %.1f", pGK.EffectiveRating())
	}

	// Two-footed LB to RB -> 0 penalty
	pBoth, _ := NewPlayer("p3", "Player 3", PosLB, FootBoth, 80.0, 25, 100.0, new(PosRB), 90.0)
	if pBoth.EffectiveRating() != 80.0 {
		t.Errorf("expected 80.0 for both-footed LB to RB, got %.1f", pBoth.EffectiveRating())
	}

	// Left-footed LB to RB -> 25% penalty
	pLeft, _ := NewPlayer("p4", "Player 4", PosLB, FootLeft, 80.0, 25, 100.0, new(PosRB), 90.0)
	if pLeft.EffectiveRating() != 60.0 {
		t.Errorf("expected 60.0 for left-footed LB to RB, got %.1f", pLeft.EffectiveRating())
	}
}

func TestHealthDeficit(t *testing.T) {
	// 90 rating with 75 health -> 65 rating
	p, err := NewPlayer("p1", "Player 1", PosCM, FootBoth, 90.0, 25, 75.0, nil,
					90.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.EffectiveRating() != 65.0 {
		t.Errorf("expected 65.0 effective rating, got %.1f", p.EffectiveRating())
	}
}

func TestTacticalEdge(t *testing.T) {
	tr := &TacticalResolver{}
	if edge := tr.GetEdge("Rock", "Scissors"); edge != 1 {
		t.Errorf("expected Rock to beat Scissors (+1), got %d", edge)
	}
	if edge := tr.GetEdge("Scissors", "Paper"); edge != 1 {
		t.Errorf("expected Scissors to beat Paper (+1), got %d", edge)
	}
	if edge := tr.GetEdge("Paper", "Rock"); edge != 1 {
		t.Errorf("expected Paper to beat Rock (+1), got %d", edge)
	}
	if edge := tr.GetEdge("Rock", "Rock"); edge != 0 {
		t.Errorf("expected Rock vs Rock to be neutral (0), got %d", edge)
	}
}

func TestLeagueMatchSimulation(t *testing.T) {
	home, err := createTestTeam("Home FC", "4-3-3 Attacking", "#ff0000", 85.0)
	if err != nil {
		t.Fatalf("failed to create home team: %v", err)
	}
	away, err := createTestTeam("Away FC", "4-2-3-1", "#0000ff", 85.0)
	if err != nil {
		t.Fatalf("failed to create away team: %v", err)
	}

	lm, err := NewLeagueMatch(home, away, true, true)
	if err != nil {
		t.Fatalf("failed to create league match: %v", err)
	}
	lm.Rng = rand.New(rand.NewSource(42))
	state := lm.Play()

	if state.Minute != 90 {
		t.Errorf("expected 90 minutes simulated, got %d", state.Minute)
	}
	if len(state.Commentary) == 0 {
		t.Errorf("expected commentary logs, got empty")
	}
	if state.HomeStats.Shots < 0 || state.AwayStats.Shots < 0 {
		t.Errorf("invalid shots count")
	}
}

func TestCupMatchSimulation(t *testing.T) {
	home, err := createTestTeam("Cup Home", "4-3-3 Attacking", "#ff0000", 85.0)
	if err != nil {
		t.Fatalf("failed to create home team: %v", err)
	}
	away, err := createTestTeam("Cup Away", "4-2-3-1", "#0000ff", 85.0)
	if err != nil {
		t.Fatalf("failed to create away team: %v", err)
	}

	cm, err := NewCupMatch(home, away, true, true)
	if err != nil {
		t.Fatalf("failed to create cup match: %v", err)
	}
	cm.Rng = rand.New(rand.NewSource(99))
	state := cm.Play()

	if state.Winner == nil {
		t.Errorf("cup match must determine a winner")
	}
}
