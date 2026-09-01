package engine

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// TournamentManager manages a standalone Knockout/Sit-and-Go tournament.
type TournamentManager struct {
	ID              string
	Name            string
	MaxParticipants int
	MaxPlayerRating float64 // Optional cap on individual player ratings (0 = no limit)
	MaxTeamRating   float64 // Optional cap on overall team rating (0 = no limit)

	WinnerRewards   map[string]int // Optional: e.g. {"Prestige": 100, "Coins": 500}
	RunnerUpRewards map[string]int // Optional: e.g. {"Prestige": 50, "Coins": 200}

	Participants []*Team
	ActiveTeams  []*Team
	Eliminated   map[string]bool

	PlayerStats map[string]*PlayerSeasonStats

	RoundNumber int
	Winner      *Team
	RunnerUp    *Team

	rng *rand.Rand
}

// NewTournamentManager initializes a custom scheduled tournament.
// MaxParticipants must be an even number (2, 4, 8, 16, etc).
func NewTournamentManager(id, name string, maxParticipants int, maxPlayerRating, maxTeamRating float64) (*TournamentManager, error) {
	if maxParticipants < 2 || maxParticipants%2 != 0 {
		return nil, fmt.Errorf("tournament MaxParticipants must be an even number (2, 4, 6, 8, etc)")
	}

	return &TournamentManager{
		ID:              id,
		Name:            name,
		MaxParticipants: maxParticipants,
		MaxPlayerRating: maxPlayerRating,
		MaxTeamRating:   maxTeamRating,
		WinnerRewards:   make(map[string]int),
		RunnerUpRewards: make(map[string]int),
		Participants:    make([]*Team, 0),
		ActiveTeams:     make([]*Team, 0),
		Eliminated:      make(map[string]bool),
		PlayerStats:     make(map[string]*PlayerSeasonStats),
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// Join attempts to add a real player's team to the tournament, applying internal simulation caps if needed.
func (tm *TournamentManager) Join(team *Team) error {
	if len(tm.Participants) >= tm.MaxParticipants {
		return fmt.Errorf("tournament is already full")
	}

	// Dynamically nerf players for the simulation instead of rejecting them
	if tm.MaxPlayerRating > 0 {
		cap := tm.MaxPlayerRating
		team.SetRatingCap(&cap)
	}

	tm.Participants = append(tm.Participants, team)
	return nil
}

// StartWithBots generates AI teams to fill any remaining empty slots so the tournament can begin on schedule.
func (tm *TournamentManager) StartWithBots() error {
	needed := tm.MaxParticipants - len(tm.Participants)

	for i := 0; i < needed; i++ {
		botRating := 60.0 // Default to 60 if no cap exists
		
		if tm.MaxPlayerRating > 0 {
			botRating = tm.MaxPlayerRating - 10.0
		} else if tm.MaxTeamRating > 0 {
			botRating = tm.MaxTeamRating - 10.0
		}

		// Pick a city name, avoiding duplicates where possible
		city := BotCities[i%len(BotCities)]
		botName := fmt.Sprintf("%s Bot FC", city)

		botTeam := GenerateBotTeam(fmt.Sprintf("%s-bot-%d", tm.ID, i+1), botName, botRating)
		tm.Participants = append(tm.Participants, botTeam)
	}

	// Copy to ActiveTeams
	tm.ActiveTeams = make([]*Team, len(tm.Participants))
	copy(tm.ActiveTeams, tm.Participants)

	// Shuffle active teams for a randomized initial bracket
	tm.rng.Shuffle(len(tm.ActiveTeams), func(i, j int) {
		tm.ActiveTeams[i], tm.ActiveTeams[j] = tm.ActiveTeams[j], tm.ActiveTeams[i]
	})

	return nil
}

// GenerateNextRound dynamically pairs the surviving teams for the next round of fixtures.
func (tm *TournamentManager) GenerateNextRound() ([]Fixture, error) {
	if tm.Winner != nil {
		return nil, fmt.Errorf("tournament is already finished")
	}

	var active []*Team
	for _, t := range tm.ActiveTeams {
		if !tm.Eliminated[t.ID] {
			active = append(active, t)
		}
	}

	if len(active) <= 1 {
		return nil, fmt.Errorf("not enough active teams to generate fixtures")
	}

	tm.RoundNumber++

	// Calculate the next lowest power of 2 for byes
	target := 1
	for target*2 <= len(active) {
		target *= 2
	}

	var fixtures []Fixture
	if len(active) == target {
		// Power of 2, everyone plays
		for i := 0; i < len(active); i += 2 {
			fixtures = append(fixtures, Fixture{Home: active[i], Away: active[i+1]})
		}
	} else {
		// Prelim round (some teams get byes)
		matchesNeeded := len(active) - target
		for i := 0; i < matchesNeeded*2; i += 2 {
			fixtures = append(fixtures, Fixture{Home: active[i], Away: active[i+1]})
		}
	}

	return fixtures, nil
}

// RecordMatch evaluates a completed Cup match, eliminates the loser, and updates Top Scorers.
func (tm *TournamentManager) RecordMatch(state *MatchState) {
	loser := state.HomeStats.Team
	winner := state.AwayStats.Team

	if state.Winner != nil && *state.Winner == "Home" {
		loser = state.AwayStats.Team
		winner = state.HomeStats.Team
	} else if state.Winner != nil && *state.Winner == "Away" {
		loser = state.HomeStats.Team
		winner = state.AwayStats.Team
	} else if state.HomeStats.GoalsFor > state.AwayStats.GoalsFor {
		loser = state.AwayStats.Team
		winner = state.HomeStats.Team
	}

	tm.Eliminated[loser.ID] = true

	// Check if this was the Final
	activeCount := 0
	for _, t := range tm.ActiveTeams {
		if !tm.Eliminated[t.ID] {
			activeCount++
		}
	}

	if activeCount == 1 {
		tm.Winner = winner
		tm.RunnerUp = loser
	}

	// Update Player Statistics
	for _, pStats := range state.PlayerStats.Stats {
		if pStats.Appearances == 0 {
			continue
		}

		ps, exists := tm.PlayerStats[pStats.Player.ID]
		if !exists {
			teamID, teamName := "", ""
			if belongsTo(pStats.Player.ID, state.HomeStats.Team) {
				teamID, teamName = state.HomeStats.Team.ID, state.HomeStats.Team.Name
			} else if belongsTo(pStats.Player.ID, state.AwayStats.Team) {
				teamID, teamName = state.AwayStats.Team.ID, state.AwayStats.Team.Name
			}

			ps = &PlayerSeasonStats{
				PlayerID:   pStats.Player.ID,
				PlayerName: pStats.Player.Name,
				TeamID:     teamID,
				TeamName:   teamName,
			}
			tm.PlayerStats[pStats.Player.ID] = ps
		}

		ps.Appearances += pStats.Appearances
		ps.Goals += pStats.Goals
		ps.Assists += pStats.Assists
		ps.CleanSheets += pStats.CleanSheets
		ps.Tackles += pStats.Tackles
	}
}

// GetTopScorers returns the top goalscorers in the tournament.
func (tm *TournamentManager) GetTopScorers(limit int) []PlayerSeasonStats {
	var list []PlayerSeasonStats
	for _, p := range tm.PlayerStats {
		if p.Goals > 0 {
			list = append(list, *p)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Goals > list[j].Goals })
	if len(list) > limit {
		return list[:limit]
	}
	return list
}

// GetWinnerID safely returns the team ID of the Tournament Winner, or empty string if not finished.
func (tm *TournamentManager) GetWinnerID() string {
	if tm.Winner != nil {
		return tm.Winner.ID
	}
	return ""
}

// GetRunnerUpID safely returns the team ID of the Tournament Runner-Up, or empty string if not finished.
func (tm *TournamentManager) GetRunnerUpID() string {
	if tm.RunnerUp != nil {
		return tm.RunnerUp.ID
	}
	return ""
}
