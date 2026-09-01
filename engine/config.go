// Package engine provides core simulation logic and configuration for the RPS Football Engine.
package engine

import (
	"fmt"
)

// Foot represents the preferred foot of a player.
type Foot string

const (
	// FootLeft indicates a player prefers their left foot.
	FootLeft Foot = "Left"
	// FootRight indicates a player prefers their right foot.
	FootRight Foot = "Right"
	// FootBoth indicates a player is ambidextrous or comfortable with both feet.
	FootBoth Foot = "Both"
)

// Global constants for player attributes and game settings.
const (
	// MinRating is the lowest possible rating for a player.
	MinRating float64 = 1.0
	// MaxRating is the highest possible rating for a player.
	MaxRating float64 = 99.0

	// MinHealth is the lowest possible health for a player.
	MinHealth float64 = 1.0
	// MaxHealth is the highest possible health for a player.
	MaxHealth float64 = 100.0

	// MinAge is the minimum age of a player.
	MinAge int = 16
	// MaxAge is the maximum age of a player before retirement.
	MaxAge int = 39

	// AgeHealthDecayThreshold is the age at which health decay becomes more severe.
	AgeHealthDecayThreshold int = 30
	// HealthDecayYoung is the health lost per match for players under the decay threshold.
	HealthDecayYoung float64 = 1.0
	// HealthDecayOld is the health lost per match for players above the decay threshold.
	HealthDecayOld float64 = 2.0

	// TacticalProgressionMod modifies the progression chance based on tactical matchups.
	TacticalProgressionMod float64 = 0.0350
	// HomeLooseBallProb is the baseline probability of the home team winning a loose ball.
	HomeLooseBallProb float64 = 0.514
	// NeutralLooseBallProb is the baseline probability of either team winning a loose ball on neutral ground.
	NeutralLooseBallProb float64 = 0.500
	// AssistedChanceProb is the probability that a goal-scoring chance is assisted.
	AssistedChanceProb float64 = 0.80
)

// DefaultPlayerRatingCap sets a global default cap on player ratings.
var DefaultPlayerRatingCap = 60.0

// Formations maps formation strings to their tactical style (Rock, Paper, or Scissors).
// This determines the tactical matchup advantage/disadvantage.
var Formations = map[string]string{
	"4-2-4":           "Rock",
	"3-4-3":           "Rock",
	"4-3-3 Attacking": "Rock",
	"4-2-3-1":         "Paper",
	"4-3-3 Holding":   "Paper",
	"3-5-2":           "Paper",
	"5-3-2":           "Scissors",
	"4-4-2 Flat":      "Scissors",
	"5-4-1":           "Scissors",
}

// FormationSlots defines the required number of players at each position for a given formation.
var FormationSlots = map[string]map[Position]int{
	// Rock formations (Aggressive)
	"4-2-4": {
		PosGK: 1, PosLB: 1, PosCB: 2, PosRB: 1,
		PosDM: 0, PosCM: 2, PosLM: 0, PosRM: 0,
		PosLW: 1, PosST: 2, PosRW: 1,
	},
	"3-4-3": {
		PosGK: 1, PosLB: 0, PosCB: 3, PosRB: 0,
		PosDM: 0, PosCM: 2, PosLM: 1, PosRM: 1,
		PosLW: 1, PosST: 1, PosRW: 1,
	},
	"4-3-3 Attacking": {
		PosGK: 1, PosLB: 1, PosCB: 2, PosRB: 1,
		PosDM: 1, PosCM: 2, PosLM: 0, PosRM: 0,
		PosLW: 1, PosST: 1, PosRW: 1,
	},
	// Paper formations (Balanced)
	"4-2-3-1": {
		PosGK: 1, PosLB: 1, PosCB: 2, PosRB: 1,
		PosDM: 2, PosCM: 1, PosLM: 1, PosRM: 1,
		PosLW: 0, PosST: 1, PosRW: 0,
	},
	"4-3-3 Holding": {
		PosGK: 1, PosLB: 1, PosCB: 2, PosRB: 1,
		PosDM: 1, PosCM: 2, PosLM: 0, PosRM: 0,
		PosLW: 1, PosST: 1, PosRW: 1,
	},
	"3-5-2": {
		PosGK: 1, PosLB: 0, PosCB: 3, PosRB: 0,
		PosDM: 1, PosCM: 2, PosLM: 1, PosRM: 1,
		PosLW: 0, PosST: 2, PosRW: 0,
	},
	// Scissors formations (Defensive)
	"5-3-2": {
		PosGK: 1, PosLB: 1, PosCB: 3, PosRB: 1,
		PosDM: 1, PosCM: 2, PosLM: 0, PosRM: 0,
		PosLW: 0, PosST: 2, PosRW: 0,
	},
	"4-4-2 Flat": {
		PosGK: 1, PosLB: 1, PosCB: 2, PosRB: 1,
		PosDM: 0, PosCM: 2, PosLM: 1, PosRM: 1,
		PosLW: 0, PosST: 2, PosRW: 0,
	},
	"5-4-1": {
		PosGK: 1, PosLB: 1, PosCB: 3, PosRB: 1,
		PosDM: 1, PosCM: 1, PosLM: 1, PosRM: 1,
		PosLW: 0, PosST: 1, PosRW: 0,
	},
}

// TacticalMatchup represents the style of the home team versus the away team.
type TacticalMatchup struct {
	Home string
	Away string
}

// RPSMatrix dictates the tactical multiplier outcome (+1 for advantage, -1 for disadvantage, 0 for neutral)
// based on Rock-Paper-Scissors rules: Rock beats Scissors, Scissors beats Paper, Paper beats Rock.
var RPSMatrix = map[TacticalMatchup]int{
	{"Rock", "Scissors"}:  1,
	{"Scissors", "Paper"}: 1,
	{"Paper", "Rock"}:     1,
	{"Scissors", "Rock"}:  -1,
	{"Paper", "Scissors"}: -1,
	{"Rock", "Paper"}:     -1,
}

// XGTier represents an Expected Goals (xG) value and its probability of occurring.
type XGTier struct {
	Value       float64
	Probability float64
}

// XGTiers determines the quality of a chance based on tactical advantage.
// A key of 1 means advantage, 0 means neutral, and -1 means disadvantage.
var XGTiers = map[int][]XGTier{
	// Advantage generates higher xG chances more frequently.
	1: {{0.38, 0.174}, {0.18, 0.407}, {0.05, 0.419}},
	// Neutral has balanced xG chances.
	0: {{0.38, 0.150}, {0.18, 0.400}, {0.05, 0.450}},
	// Disadvantage leads to lower quality chances overall.
	-1: {{0.38, 0.133}, {0.18, 0.393}, {0.05, 0.474}},
}

// CalculateOOPPenalty computes a fractional penalty [0.0 to 0.5] applied to a player's
// rating if they are playing out of their natural position.
func CalculateOOPPenalty(natural, assigned Position, foot Foot) float64 {
	// No penalty if playing in their natural position.
	if natural == assigned {
		return 0.0
	}
	// Massive 50% penalty for outfielders in goal, or goalkeepers outfield.
	if natural == PosGK || assigned == PosGK {
		return 0.50
	}

	// Two-footed players have no penalty when switching flanks (e.g. LB to RB).
	isFullbackSwap := (natural == PosLB && assigned == PosRB) || (natural == PosRB && assigned == PosLB)
	isWideMidSwap := (natural == PosLM && assigned == PosRM) || (natural == PosRM && assigned == PosLM)
	isWingerSwap := (natural == PosLW && assigned == PosRW) || (natural == PosRW && assigned == PosLW)
	if (isFullbackSwap || isWideMidSwap || isWingerSwap) && foot == FootBoth {
		return 0.0
	}

	// Helper functions to categorize positional lines.
	isDef := func(p Position) bool { return p == PosLB || p == PosCB || p == PosRB }
	isMid := func(p Position) bool { return p == PosLM || p == PosDM || p == PosCM || p == PosRM }
	isFwd := func(p Position) bool { return p == PosLW || p == PosST || p == PosRW }

	// 25% penalty for shifting horizontally within the same line (e.g. CB to RB, CM to LM).
	if (isDef(natural) && isDef(assigned)) ||
		(isMid(natural) && isMid(assigned)) ||
		(isFwd(natural) && isFwd(assigned)) {
		return 0.25
	}

	// 30% penalty for moving up/down within the same vertical channel
	// (e.g. Left Back to Left Mid, Center Back to Defensive Mid).
	channelPairs := map[string]bool{
		"LB->LM": true, "LM->LB": true, "RB->RM": true, "RM->RB": true,
		"LW->LM": true, "LM->LW": true, "RW->RM": true, "RM->RW": true,
		"CB->DM": true, "DM->CB": true, "CM->ST": true, "ST->CM": true,
	}
	if channelPairs[fmt.Sprintf("%s->%s", natural, assigned)] {
		return 0.30
	}

	// 35% penalty for moving to adjacent lines not in the same channel (e.g. Def to Mid, or Mid to Fwd).
	if (isDef(natural) && isMid(assigned)) || (isMid(natural) && isDef(assigned)) ||
		(isMid(natural) && isFwd(assigned)) || (isFwd(natural) && isMid(assigned)) {
		return 0.35
	}

	// 45% penalty for skipping a line entirely (e.g. Defender to Forward).
	if (isDef(natural) && isFwd(assigned)) || (isFwd(natural) && isDef(assigned)) {
		return 0.45
	}

	// Fallback maximum penalty.
	return 0.50
}
