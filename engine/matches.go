package engine
// Package engine provides match orchestration for the RPS Football Engine.

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// ApplyPostMatchFatigue calculates and applies health degradation for all players
// after a match. Older players (above AgeHealthDecayThreshold) lose more health.
// The amount lost is recorded in the MatchState's PlayerStats.
func ApplyPostMatchFatigue(state *MatchState, home, away *Team) {
	for _, p := range append(home.Players, away.Players...) {
		decay := HealthDecayYoung
		if p.Age > AgeHealthDecayThreshold {
			decay = HealthDecayOld
		}

		actualLost := decay
		if p.Health-decay < MinHealth {
			actualLost = p.Health - MinHealth
		}

		p.Health -= max(0, actualLost)
		entry := state.PlayerStats.GetOrCreate(p)
		entry.MatchHealthLost = max(0, actualLost)
		entry.PostMatchHealth = p.Health
	}
}

// CoinTossKickoff randomly determines which team gets initial possession
// at the start of a match.
func CoinTossKickoff(state *MatchState, rng *rand.Rand) {
	if rng.Float64() < 0.50 {
		state.PossessionTeam = "Home"
	} else {
		state.PossessionTeam = "Away"
	}
}

// RecordStartingMatches logs an appearance for all players in the starting
// lineups for both teams in the match stats.
func RecordStartingMatches(state *MatchState, home, away *Team) {
	for _, p := range append(home.Players, away.Players...) {
		state.PlayerStats.RecordMatchPlayed(p)
	}
}

// LeagueMatch represents a standard league game. League games only last 90 minutes
// and can end in a draw. No extra time or penalty shootout is played.
type LeagueMatch struct {
	Home    *Team      // The home team
	Away    *Team      // The away team
	HomeAdv bool       // Whether home advantage is applied
	Verbose bool       // Whether to print detailed match events
	Rng     *rand.Rand // Random number generator for match events
}

// NewLeagueMatch creates a new LeagueMatch, validating lineups before proceeding.
func NewLeagueMatch(home, away *Team, homeAdv, verbose bool) (*LeagueMatch, error) {
	if err := home.ValidateLineup(); err != nil {
		return nil, err
	}
	if err := away.ValidateLineup(); err != nil {
		return nil, err
	}
	return &LeagueMatch{
		Home:    home,
		Away:    away,
		HomeAdv: homeAdv,
		Verbose: verbose,
		Rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// Play simulates a LeagueMatch. It initializes the MatchState, plays out 90 minutes,
// determines the winner (if any), and records final statistics like clean sheets and fatigue.
func (m *LeagueMatch) Play() *MatchState {
	// Initialize the MatchState specific for a League match
	state := NewMatchState(MatchLeague, m.Home, m.Away)
	CoinTossKickoff(state, m.Rng)
	RecordStartingMatches(state, m.Home, m.Away)

	if m.Verbose {
		state.Log(fmt.Sprintf("Formation: %s ( %s ) : %s ( %s )", m.Home.Formation, m.Home.Style(), m.Away.Formation, m.Away.Style()))
		state.Log("Kick Off!")
	}

	engine := NewTickEngine(m.Home, m.Away, m.HomeAdv, m.Verbose, m.Rng)
	engine.ExecuteTicks(state, 1, 90)

	hStr, aStr := "Home", "Away"
	if state.HomeStats.MatchGoalsFor > state.AwayStats.MatchGoalsFor {
		state.Winner = &hStr
		state.HomeStats.MatchWins = 1
		state.AwayStats.MatchLosses = 1
	} else if state.AwayStats.MatchGoalsFor > state.HomeStats.MatchGoalsFor {
		state.Winner = &aStr
		state.AwayStats.MatchWins = 1
		state.HomeStats.MatchLosses = 1
	} else {
		state.Winner = nil
		state.HomeStats.MatchDraws = 1
		state.AwayStats.MatchDraws = 1
	}

	if m.Verbose {
		LogPostMatchStats(state)
	}

	state.PlayerStats.RecordCleanSheets(m.Home, m.Away, state.HomeStats.MatchGoalsFor, state.AwayStats.MatchGoalsFor)
	ApplyPostMatchFatigue(state, m.Home, m.Away)
	FinalizeClubStats(state)
	return state
}

// CupMatch represents a knockout tournament match. If the score is tied after 90 minutes,
// it proceeds to extra time (30 minutes). If still tied, it is decided by a penalty shootout.
type CupMatch struct {
	Home    *Team      // The home team
	Away    *Team      // The away team
	HomeAdv bool       // Whether home advantage is applied
	Verbose bool       // Whether to print detailed match events
	Rng     *rand.Rand // Random number generator for match events
}

// NewCupMatch creates a new CupMatch, validating lineups before proceeding.
func NewCupMatch(home, away *Team, homeAdv, verbose bool) (*CupMatch, error) {
	if err := home.ValidateLineup(); err != nil {
		return nil, err
	}
	if err := away.ValidateLineup(); err != nil {
		return nil, err
	}
	return &CupMatch{
		Home:    home,
		Away:    away,
		HomeAdv: homeAdv,
		Verbose: verbose,
		Rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// Play simulates a CupMatch. It handles regulation time, extra time if tied, and
// penalty shootouts if the tie persists. MatchState and stats are tracked throughout.
func (m *CupMatch) Play() *MatchState {
	// Initialize the MatchState specific for a Cup match
	state := NewMatchState(MatchCup, m.Home, m.Away)
	CoinTossKickoff(state, m.Rng)
	RecordStartingMatches(state, m.Home, m.Away)

	if m.Verbose {
		state.Log(fmt.Sprintf("Formation: %s ( %s ) : %s ( %s )", m.Home.Formation, m.Home.Style(), m.Away.Formation, m.Away.Style()))
		state.Log("Kick Off!")
	}

	engine := NewTickEngine(m.Home, m.Away, m.HomeAdv, m.Verbose, m.Rng)
	engine.ExecuteTicks(state, 1, 90)

	if state.HomeStats.MatchGoalsFor == state.AwayStats.MatchGoalsFor {
		state.WentToExtraTime = true
		if m.Verbose {
			state.Log("--- EXTRA TIME BEGINS (91'-120') ---")
		}
		state.BallZone = ZoneMidfield
		state.PossessionTeam = "Home"
		engine.ExecuteTicks(state, 91, 120)
	}

	hStr, aStr := "Home", "Away"

	if state.HomeStats.MatchGoalsFor == state.AwayStats.MatchGoalsFor {
		penResolver := &PenaltyResolver{}
		penResolver.ResolveShootout(state, m.Home, m.Away, m.Verbose, m.Rng)
	} else {
		if state.HomeStats.MatchGoalsFor > state.AwayStats.MatchGoalsFor {
			state.Winner = &hStr
		} else {
			state.Winner = &aStr
		}
	}

	if state.Winner != nil {
		if *state.Winner == "Home" {
			state.HomeStats.MatchWins = 1
			state.AwayStats.MatchLosses = 1
		} else {
			state.AwayStats.MatchWins = 1
			state.HomeStats.MatchLosses = 1
		}
	}

	if m.Verbose {
		LogPostMatchStats(state)
	}

	state.PlayerStats.RecordCleanSheets(m.Home, m.Away, state.HomeStats.MatchGoalsFor, state.AwayStats.MatchGoalsFor)
	ApplyPostMatchFatigue(state, m.Home, m.Away)
	FinalizeClubStats(state)
	return state
}

// LogPostMatchStats appends final match summary statistics with their values
func LogPostMatchStats(state *MatchState) {
	state.Log("The Final Whistle!")

	totalTicks := state.HomeStats.PossessionTicks + state.AwayStats.PossessionTicks
	homePoss := 50
	awayPoss := 50
	if totalTicks > 0 {
		homePoss = int(math.Round(float64(state.HomeStats.PossessionTicks) / float64(totalTicks) * 100.0))
		awayPoss = 100 - homePoss
	}

	state.Log(fmt.Sprintf("Expected Goals (xG): %.2f : %.2f", state.HomeStats.XG, state.AwayStats.XG))
	state.Log(fmt.Sprintf("Possession: %d%% : %d%%", homePoss, awayPoss))
	state.Log(fmt.Sprintf("Shots: %d : %d", state.HomeStats.Shots, state.AwayStats.Shots))
	state.Log(fmt.Sprintf("Shots On Target: %d : %d", state.HomeStats.SOT, state.AwayStats.SOT))
	state.Log(fmt.Sprintf("Saves: %d : %d", state.HomeStats.Saves, state.AwayStats.Saves))
	state.Log(fmt.Sprintf("Duels Won: %d : %d", state.HomeStats.Tackles, state.AwayStats.Tackles))
	state.Log(fmt.Sprintf("Team Rating: %.0f : %.0f", state.HomeStats.TeamRating, state.AwayStats.TeamRating))
}

// QuickPlay is a developer-friendly facade that automatically initializes and plays a match.
// It handles RNG injection and returns the completed MatchState delta payload.
func QuickPlay(matchType MatchType, home, away *Team, homeAdv bool) (*MatchState, error) {
	if matchType == MatchLeague {
		match, err := NewLeagueMatch(home, away, homeAdv, true) // verbose=true to generate commentary logs
		if err != nil {
			return nil, err
		}
		return match.Play(), nil
	} else if matchType == MatchCup {
		match, err := NewCupMatch(home, away, homeAdv, true)
		if err != nil {
			return nil, err
		}
		return match.Play(), nil
	}
	return nil, fmt.Errorf("invalid match type: %s", matchType)
}

// DeterministicPlay operates identically to QuickPlay, but accepts a specific RNG seed.
// Passing the same seed with the same teams will always generate the exact same match events and scoreline.
func DeterministicPlay(matchType MatchType, home, away *Team, homeAdv bool, seed int64) (*MatchState, error) {
	if matchType == MatchLeague {
		match, err := NewLeagueMatch(home, away, homeAdv, true)
		if err != nil {
			return nil, err
		}
		match.Rng = rand.New(rand.NewSource(seed))
		return match.Play(), nil
	} else if matchType == MatchCup {
		match, err := NewCupMatch(home, away, homeAdv, true)
		if err != nil {
			return nil, err
		}
		match.Rng = rand.New(rand.NewSource(seed))
		return match.Play(), nil
	}
	return nil, fmt.Errorf("invalid match type: %s", matchType)
}

func FinalizeClubStats(state *MatchState) {
    if state.Winner == nil && state.MatchType == MatchLeague {
        state.HomeStats.MatchDraws = 1
        state.AwayStats.MatchDraws = 1
    }

    updateClub(state.HomeStats, state.AwayStats)
    updateClub(state.AwayStats, state.HomeStats)
}

func updateClub(stats *ClubMatchStats, oppStats *ClubMatchStats) {
    team := stats.Team
    team.Matches += stats.MatchMatches
    team.Wins += stats.MatchWins
    team.Draws += stats.MatchDraws
    team.Losses += stats.MatchLosses
    team.GoalsFor += stats.MatchGoalsFor
    team.GoalsAgainst += oppStats.MatchGoalsFor

    stats.MatchGoalsAgainst = oppStats.MatchGoalsFor
    
    stats.PostMatchMatches = team.Matches
    stats.PostMatchWins = team.Wins
    stats.PostMatchDraws = team.Draws
    stats.PostMatchLosses = team.Losses
    stats.PostMatchGoalsFor = team.GoalsFor
    stats.PostMatchGoalsAgainst = team.GoalsAgainst
}
