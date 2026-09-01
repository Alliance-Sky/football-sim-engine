// Package engine provides core simulation logic for the RPS Football Engine.
package engine

import (
	"fmt"
	"math/rand"
)

// TickEngine drives the core match loop.
// It iterates tick-by-tick (minute-by-minute) resolving team actions:
// progressing the ball, holding possession, turning it over, or taking a shot.
type TickEngine struct {
	Home                *Team
	Away                *Team
	Verbose             bool
	TacticalEdge        int
	ProgressionResolver *ProgressionResolver
	ShotResolver        *ShotResolver
	Rng                 *rand.Rand
}

// NewTickEngine initializes a new engine, setting up sub-resolvers (progression, shots, tactics)
// based on the teams' strengths, styles, and home field advantage.
func NewTickEngine(home, away *Team, homeAdv, verbose bool, rng *rand.Rand) *TickEngine {
	tacticalResolver := &TacticalResolver{}
	// Calculate Rock-Paper-Scissors tactical edge based on team styles
	edge := tacticalResolver.GetEdge(home.Style(), away.Style())
	return &TickEngine{
		Home:                home,
		Away:                away,
		Verbose:             verbose,
		TacticalEdge:        edge,
		ProgressionResolver: NewProgressionResolver(home, away, edge, homeAdv, verbose, rng),
		ShotResolver:        NewShotResolver(home, away, edge, homeAdv, verbose, rng),
		Rng:                 rng,
	}
}

// ExecuteTicks runs the simulation from startTick to endTick.
// A standard match runs from tick 1 to 90.
// Each tick (minute) the team in possession attempts to advance the ball towards the opponent's goal.
func (te *TickEngine) ExecuteTicks(state *MatchState, startTick, endTick int) {
	for tick := startTick; tick <= endTick; tick++ {
		state.Minute = tick // 1 tick = 1 minute of match time
		isHome := (state.PossessionTeam == "Home")

		if isHome {
			state.HomeStats.PossessionTicks++
		} else {
			state.AwayStats.PossessionTicks++
		}

		// Identify the attacking target zone
		targetGoalZone := ZoneAwayBox
		if !isHome {
			targetGoalZone = ZoneHomeBox
		}

		// If the ball is already in the opponent's box, attempt a shot
		if state.BallZone == targetGoalZone {
			te.ShotResolver.Resolve(state, isHome)
			continue
		}

		// Otherwise, attempt to progress the ball
		pProg := te.ProgressionResolver.CalculateProgressionProb(isHome)
		roll := te.Rng.Float64()

		if roll < pProg {
			// Successful progression - move the ball forward one zone
			if isHome {
				state.BallZone++
			} else {
				state.BallZone--
			}
			if te.Verbose && ((isHome && state.BallZone == ZoneAwayHalf) || (!isHome && state.BallZone == ZoneHomeHalf)) && te.Rng.Float64() < 0.22 {
				attTeam := te.Home
				if !isHome {
					attTeam = te.Away
				}
				progressors := append(attTeam.Strikers(), attTeam.Wingers()...)
				if len(progressors) > 0 {
					p := progressors[te.Rng.Intn(len(progressors))]
					state.Log(fmt.Sprintf("▲▲▲ Counterattack - %s", p.Name))
				}
			}
		} else if roll < pProg+0.25 {
			// Held possession - ball stays in the same zone, no turnover
		} else {
			// Turnover - the defending team wins the ball
			state.PossessionTeam = te.ProgressionResolver.ResolveTurnover(state, state.BallZone, isHome)
		}
	}
}
