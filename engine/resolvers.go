// Package engine provides core simulation logic and sub-resolvers for the RPS Football Engine.
package engine

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

func weightedChoice[T any](items []T, weights []float64, r *rand.Rand) T {
	var total float64
	for _, w := range weights {
		total += w
	}
	roll := r.Float64() * total
	var cumulative float64
	for i, w := range weights {
		cumulative += w
		if roll <= cumulative || i == len(items)-1 {
			return items[i]
		}
	}
	return items[len(items)-1]
}

// TacticalResolver determines if one team has a tactical advantage over the other.
// It uses a Rock-Paper-Scissors mechanic based on the teams' playstyles.
type TacticalResolver struct{}

// GetEdge returns the tactical advantage value (from the RPSMatrix).
// For example, if Home style counters Away style (e.g., Counter Attack beats Possession),
// Home gets a positive edge. If they are countered, they get a negative edge.
func (tr *TacticalResolver) GetEdge(homeStyle, awayStyle string) int {
	return RPSMatrix[TacticalMatchup{Home: homeStyle, Away: awayStyle}]
}

// ProgressionResolver handles the logic for moving the ball up or down the pitch.
// It uses team ratings, home advantage, and the tactical edge to determine
// the likelihood of a team successfully advancing the ball towards the goal.
type ProgressionResolver struct {
	Home          *Team
	Away          *Team
	TacticalEdge  int
	LooseBallProb float64
	Verbose       bool
	Rng           *rand.Rand
}

func NewProgressionResolver(home, away *Team, tacticalEdge int, homeAdv, verbose bool, rng *rand.Rand) *ProgressionResolver {
	prob := NeutralLooseBallProb
	if homeAdv {
		prob = HomeLooseBallProb
	}
	return &ProgressionResolver{
		Home:          home,
		Away:          away,
		TacticalEdge:  tacticalEdge,
		LooseBallProb: prob,
		Verbose:       verbose,
		Rng:           rng,
	}
}

func (pr *ProgressionResolver) CalculateProgressionProb(isHome bool) float64 {
	var rAtt, rDef float64
	if isHome {
		rAtt = pr.Home.ProgressionRating()
		rDef = pr.Away.ProgressionRating()
	} else {
		rAtt = pr.Away.ProgressionRating()
		rDef = pr.Home.ProgressionRating()
	}

	baseP := 0.50 + ((rAtt - rDef) / 300.0)
	tacticalMod := TacticalProgressionMod * float64(pr.TacticalEdge)
	if !isHome {
		tacticalMod = -tacticalMod
	}
	return max(0.20, min(0.80, baseP+tacticalMod))
}

func (pr *ProgressionResolver) ResolveTurnover(state *MatchState, currentZone PitchZone, isHome bool) string {
	nextPossession := "Away"
	if pr.Rng.Float64() < pr.LooseBallProb {
		nextPossession = "Home"
	}

	isFinalThird := (isHome && currentZone == ZoneAwayHalf) || (!isHome && currentZone == ZoneHomeHalf)
	if state != nil && isFinalThird {
		defTeam := pr.Away
		if !isHome {
			defTeam = pr.Home
		}

		candidates := append(append(defTeam.Defenders(), defTeam.DefensiveMids()...), defTeam.CentralMids()...)
		if len(candidates) > 0 {
			weights := make([]float64, len(candidates))
			tackleWeights := map[Position]float64{
				PosCB: 10.0, PosDM: 8.0, PosLB: 5.0, PosRB: 5.0, PosCM: 2.0, PosLM: 1.0, PosRM: 1.0,
			}
			for i, p := range candidates {
				baseWeight := tackleWeights[p.AssignedPosition]
				if baseWeight == 0 {
					baseWeight = 1.0
				}
				weights[i] = baseWeight * (p.EffectiveRating() / 60.0)
			}

			defender := weightedChoice(candidates, weights, pr.Rng)
			if isHome {
				state.AwayStats.Tackles++
			} else {
				state.HomeStats.Tackles++
			}

			if defender.AssignedPosition.IsDefender() || defender.AssignedPosition.IsDefensiveMid() {
				state.PlayerStats.RecordTackle(defender)
			}

			if pr.Verbose && pr.Rng.Float64() < 0.28 {
				state.Log(fmt.Sprintf("🛑 %d. %s - tackle", state.Minute, defender.Name))
			}
		}
	}
	return nextPossession
}

// ShotResolver handles the outcome of a goal-scoring opportunity once the ball
// reaches the opponent's box. It decides whether the chance comes from open play,
// a free kick, or a corner, assigns expected goals (xG) based on team ratings and
// tactical edge, picks a shooter, and resolves if it's a goal, save, or miss.
type ShotResolver struct {
	Home          *Team
	Away          *Team
	TacticalEdge  int
	LooseBallProb float64
	Verbose       bool
	Rng           *rand.Rand
}

func NewShotResolver(home, away *Team, tacticalEdge int, homeAdv, verbose bool, rng *rand.Rand) *ShotResolver {
	prob := NeutralLooseBallProb
	if homeAdv {
		prob = HomeLooseBallProb
	}
	return &ShotResolver{
		Home:          home,
		Away:          away,
		TacticalEdge:  tacticalEdge,
		LooseBallProb: prob,
		Verbose:       verbose,
		Rng:           rng,
	}
}

func (sr *ShotResolver) sampleXGTier(edge int, attTeam, defTeam *Team) (float64, string) {
	baseTiers := XGTiers[edge]
	creationDelta := (attTeam.CreationThreatRating() - defTeam.DefensiveShieldRating()) / 400.0
	creationDelta = max(-0.04, min(0.04, creationDelta))

	roll := sr.Rng.Float64()
	var cumulative float64

	for idx, tier := range baseTiers {
		adjustedProb := tier.Probability
		if tier.Value == 0.38 {
			adjustedProb = max(0.05, tier.Probability+creationDelta)
		} else if tier.Value == 0.05 {
			adjustedProb = max(0.05, tier.Probability-creationDelta)
		}

		cumulative += adjustedProb
		if roll <= cumulative || idx == len(baseTiers)-1 {
			var label string
			switch tier.Value {
			case 0.38:
				label = "High Danger Overload"
			case 0.18:
				label = "Contested Box Chance"
			default:
				label = "Low Danger Effort"
			}
			return tier.Value, label
		}
	}
	return 0.05, "Low Danger Effort"
}

func (sr *ShotResolver) selectShooter(attTeam *Team, isCorner bool) *Player {
	shotVolume := map[Position]float64{
		PosST: 10.0, PosLW: 4.5, PosRW: 4.5, PosCM: 2.0, PosLM: 1.5, PosRM: 1.5,
		PosDM: 0.8, PosLB: 0.4, PosRB: 0.4, PosCB: 0.3, PosGK: 0.0,
	}
	cornerShooters := map[Position]float64{
		PosCB: 9.0, PosST: 8.0, PosDM: 3.5, PosCM: 2.0, PosLW: 1.5, PosRW: 1.5,
		PosLB: 0.8, PosRB: 0.8, PosLM: 0.5, PosRM: 0.5, PosGK: 0.0,
	}

	weights := make([]float64, len(attTeam.Players))
	for i, p := range attTeam.Players {
		if isCorner {
			weights[i] = cornerShooters[p.AssignedPosition]
		} else {
			weights[i] = shotVolume[p.AssignedPosition]
		}
	}
	return weightedChoice(attTeam.Players, weights, sr.Rng)
}

func (sr *ShotResolver) selectCreator(attTeam *Team, shooter *Player, isCorner bool) *Player {
	if !isCorner && sr.Rng.Float64() > AssistedChanceProb {
		return nil
	}

	var candidates []*Player
	for _, p := range attTeam.Players {
		if p != shooter && !p.AssignedPosition.IsGoalkeeper() {
			candidates = append(candidates, p)
		}
	}
	if len(candidates) == 0 {
		return nil
	}

	weights := make([]float64, len(candidates))
	cornerWeights := map[Position]float64{
		PosCM: 8.0, PosLM: 7.0, PosRM: 7.0, PosLW: 7.0, PosRW: 7.0, PosLB: 4.0, PosRB: 4.0,
	}
	openWeights := map[Position]float64{
		PosCM: 8.0, PosLW: 7.0, PosRW: 7.0, PosLM: 5.0, PosRM: 5.0,
		PosLB: 3.0, PosRB: 3.0, PosST: 2.5, PosDM: 2.0, PosCB: 0.5,
	}

	for i, p := range candidates {
		var w float64
		if isCorner {
			w = cornerWeights[p.AssignedPosition]
		} else {
			w = openWeights[p.AssignedPosition]
		}
		if w == 0 {
			w = 0.1
		}
		weights[i] = max(0.1, w*(p.EffectiveRating()/60.0))
	}
	return weightedChoice(candidates, weights, sr.Rng)
}

func (sr *ShotResolver) Resolve(state *MatchState, isHome bool) {
	if isHome {
		state.HomeStats.Shots++
	} else {
		state.AwayStats.Shots++
	}

	edge := sr.TacticalEdge
	attTeam, defTeam := sr.Home, sr.Away
	if !isHome {
		edge = -edge
		attTeam, defTeam = sr.Away, sr.Home
	}

	chanceRoll := sr.Rng.Float64()
	var isCorner bool

	if chanceRoll < 0.14 {
		isCorner = true
	}

	selectedXG, _ := sr.sampleXGTier(edge, attTeam, defTeam)
	if isHome {
		state.HomeStats.XG += selectedXG
	} else {
		state.AwayStats.XG += selectedXG
	}
	shooter := sr.selectShooter(attTeam, isCorner)
	creator := sr.selectCreator(attTeam, shooter, isCorner)

	gk := defTeam.Goalkeepers()[0]
	defUnitRating := max(MinRating, defTeam.DefenseUnitRating())
	shooterRating := max(MinRating, shooter.EffectiveRating())

	goalProb := min(0.95, selectedXG*math.Pow(shooterRating/defUnitRating, 0.75))
	roll := sr.Rng.Float64()

	if roll < goalProb {
		state.PlayerStats.RecordGoal(shooter)
		if creator != nil {
			state.PlayerStats.RecordAssist(creator)
		}

		if isHome {
			state.HomeStats.GoalsFor++
			state.AwayStats.GoalsAgainst++
			state.HomeStats.SOT++
			if creator != nil {
				state.HomeStats.Assists++
			}
		} else {
			state.AwayStats.GoalsFor++
			state.HomeStats.GoalsAgainst++
			state.AwayStats.SOT++
			if creator != nil {
				state.AwayStats.Assists++
			}
		}

		if sr.Verbose {
			state.Log(fmt.Sprintf("⚽ GOAL! %d. %d:%d %s",
				state.Minute, state.HomeStats.GoalsFor, state.AwayStats.GoalsFor, shooter.Name))
		}

		state.BallZone = ZoneMidfield
		if isHome {
			state.PossessionTeam = "Away"
		} else {
			state.PossessionTeam = "Home"
		}
	} else {
		margin := roll - goalProb
		if margin < 0.035 {
			if sr.Rng.Float64() < 0.50 {
				if sr.Verbose {
					woodwork := "post"
					if sr.Rng.Float64() < 0.50 {
						woodwork = "crossbar"
					}
					state.Log(fmt.Sprintf("💥 %d. %s - %s",
						state.Minute, shooter.Name, woodwork))
				}
			} else {
				if isHome {
					state.HomeStats.SOT++
				} else {
					state.AwayStats.SOT++
				}
				defender := defTeam.Defenders()[sr.Rng.Intn(len(defTeam.Defenders()))]
				if sr.Verbose {
					state.Log(fmt.Sprintf("🛡️ %d. %s - clearance",
						state.Minute, defender.Name))
				}
			}
		} else {
			if sr.Rng.Float64() < 0.65 {
				if isHome {
					state.HomeStats.SOT++
					state.AwayStats.Saves++
				} else {
					state.AwayStats.SOT++
					state.HomeStats.Saves++
				}
				if sr.Verbose {
					state.Log(fmt.Sprintf("🧤 %d. %s - save",
						state.Minute, gk.Name))
				}
			} else if sr.Verbose {
				state.Log(fmt.Sprintf("⚽ %d. %s - miss",
					state.Minute, shooter.Name))
			}
		}

		if sr.Rng.Float64() < sr.LooseBallProb {
			state.PossessionTeam = "Home"
		} else {
			state.PossessionTeam = "Away"
		}

		if (isHome && state.PossessionTeam == "Away") || (!isHome && state.PossessionTeam == "Home") {
			if state.PossessionTeam == "Home" {
				state.BallZone = ZoneHomeHalf
			} else {
				state.BallZone = ZoneAwayHalf
			}
		}
	}
}

type PenaltyResolver struct{}

func (pr *PenaltyResolver) Kick(shooter, gk *Player, rng *rand.Rand) bool {
	sRating := max(MinRating, shooter.EffectiveRating())
	gRating := max(MinRating, gk.EffectiveRating())
	pScore := max(0.40, min(0.95, 0.75*math.Pow(sRating/gRating, 0.8)))
	return rng.Float64() < pScore
}

func (pr *PenaltyResolver) ResolveShootout(state *MatchState, home, away *Team, verbose bool, rng *rand.Rand) {
	state.WentToPenalties = true
	if verbose {
		state.Log("--- PENALTY SHOOTOUT STARTS ---")
	}

	orderKey := map[Position]int{
		PosST: 1, PosLW: 2, PosRW: 2, PosCM: 3, PosLM: 3, PosRM: 3,
		PosDM: 4, PosLB: 5, PosRB: 5, PosCB: 6, PosGK: 7,
	}

	sortTakers := func(t *Team) []*Player {
		sorted := make([]*Player, len(t.Players))
		copy(sorted, t.Players)
		sort.SliceStable(sorted, func(i, j int) bool {
			ki, kj := orderKey[sorted[i].AssignedPosition], orderKey[sorted[j].AssignedPosition]
			if ki != kj {
				return ki < kj
			}
			return sorted[i].EffectiveRating() > sorted[j].EffectiveRating()
		})
		return sorted
	}

	homeTakers := sortTakers(home)
	awayTakers := sortTakers(away)
	homeGK := home.Goalkeepers()[0]
	awayGK := away.Goalkeepers()[0]

	homeScored, awayScored := 0, 0

	for r := 1; r <= 5; r++ {
		hTaker := homeTakers[(r-1)%len(homeTakers)]
		if pr.Kick(hTaker, awayGK, rng) {
			homeScored++
			entry := fmt.Sprintf("⚽ Penalty Round %d | %s - scored", r, hTaker.Name)
			state.PenaltyShootoutLog = append(state.PenaltyShootoutLog, entry)
			if verbose {
				state.Log(entry)
			}
		} else {
			entry := fmt.Sprintf("🧤 Penalty Round %d | %s - miss", r, hTaker.Name)
			state.PenaltyShootoutLog = append(state.PenaltyShootoutLog, entry)
			if verbose {
				state.Log(entry)
			}
		}

		hRem, aRem := 5-r, 5-(r-1)
		if homeScored > awayScored+aRem || awayScored > homeScored+hRem {
			break
		}

		aTaker := awayTakers[(r-1)%len(awayTakers)]
		if pr.Kick(aTaker, homeGK, rng) {
			awayScored++
			entry := fmt.Sprintf("⚽ Penalty Round %d | %s - scored", r, aTaker.Name)
			state.PenaltyShootoutLog = append(state.PenaltyShootoutLog, entry)
			if verbose {
				state.Log(entry)
			}
		} else {
			entry := fmt.Sprintf("🧤 Penalty Round %d | %s - miss", r, aTaker.Name)
			state.PenaltyShootoutLog = append(state.PenaltyShootoutLog, entry)
			if verbose {
				state.Log(entry)
			}
		}

		aRem = 5 - r
		if homeScored > awayScored+aRem || awayScored > homeScored+hRem {
			break
		}
	}

	sd := 6
	for homeScored == awayScored {
		if verbose && sd == 6 {
			state.Log("--- SUDDEN DEATH PENALTIES ---")
		}
		hTaker := homeTakers[(sd-1)%len(homeTakers)]
		aTaker := awayTakers[(sd-1)%len(awayTakers)]

		hHit := pr.Kick(hTaker, awayGK, rng)
		aHit := pr.Kick(aTaker, homeGK, rng)

		if hHit {
			homeScored++
		}
		if aHit {
			awayScored++
		}

		hRes, aRes := "miss", "miss"
		hIcon, aIcon := "🧤", "🧤"
		if hHit {
			hRes = "scored"
			hIcon = "⚽"
		}
		if aHit {
			aRes = "scored"
			aIcon = "⚽"
		}

		entryH := fmt.Sprintf("%s Sudden Death %d | %s - %s", hIcon, sd, hTaker.Name, hRes)
		entryA := fmt.Sprintf("%s Sudden Death %d | %s - %s", aIcon, sd, aTaker.Name, aRes)
		state.PenaltyShootoutLog = append(state.PenaltyShootoutLog, entryH, entryA)
		if verbose {
			state.Log(entryH)
			state.Log(entryA)
		}
		sd++
	}

	hStr, aStr := "Home", "Away"
	if homeScored > awayScored {
		state.Winner = &hStr
	} else {
		state.Winner = &aStr
	}
}
