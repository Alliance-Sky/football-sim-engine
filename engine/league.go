package engine

import (
	"fmt"
	"sort"
)

// Fixture represents a scheduled match that hasn't been played yet.
type Fixture struct {
	Home *Team `json:"home"`
	Away *Team `json:"away"`
}

// TableStanding represents a single team's current rank and stats in the league.
type TableStanding struct {
	Rank           int      `json:"rank"`
	TeamID         string   `json:"teamId"`
	TeamName       string   `json:"teamName"`
	Played         int      `json:"played"`
	Wins           int      `json:"wins"`
	Draws          int      `json:"draws"`
	Losses         int      `json:"losses"`
	GoalsFor       int      `json:"goalsFor"`
	GoalsAgainst   int      `json:"goalsAgainst"`
	GoalDifference int      `json:"goalDifference"`
	Points         int      `json:"points"`
	Form           []string `json:"form"`
}

// PlayerSeasonStats tracks aggregated stats for a player across the entire season.
type PlayerSeasonStats struct {
	PlayerID    string `json:"playerId"`
	PlayerName  string `json:"playerName"`
	TeamID      string `json:"teamId"`
	TeamName    string `json:"teamName"`
	Appearances int    `json:"appearances"`
	Goals       int    `json:"goals"`
	Assists     int    `json:"assists"`
	CleanSheets int    `json:"cleanSheets"`
	Tackles     int    `json:"tackles"`
}

// ScheduledRound represents a single matchday, which could be either a League matchday or a Cup knockout round.
type ScheduledRound struct {
	Type     MatchType
	Fixtures []Fixture
}

// LeagueManager tracks scheduling, team standings, and player statistics for a full season,
// including an interleaved knockout League Cup competition.
type LeagueManager struct {
	Teams        []*Team
	Schedule     []ScheduledRound
	CurrentRound int

	MaxPlayerRating float64 // Optional cap (0 = no limit)
	MaxTeamRating   float64 // Optional cap (0 = no limit)

	// PositionRewards map leaderboard rank (1, 2, 3...) to a dictionary of custom prizes
	PositionRewards map[int]map[string]int

	Standings     map[string]*TableStanding
	PlayerStats   map[string]*PlayerSeasonStats
	CupEliminated map[string]bool
}

// NewLeagueManager initializes a new league, builds the standings map, and generates the interleaved round-robin and cup schedule.
// If maxPlayerRating is > 0, it dynamically caps all players in the simulation.
// roundRobinMultiplier determines how many times teams play each other (1 = Single, 2 = Double, 3 = Triple, etc).
func NewLeagueManager(teams []*Team, maxPlayerRating, maxTeamRating float64, roundRobinMultiplier int) (*LeagueManager, error) {
	if len(teams) < 2 {
		return nil, fmt.Errorf("league requires at least 2 teams")
	}

	// Dynamically nerf players for the simulation instead of rejecting them
	if maxPlayerRating > 0 {
		for _, team := range teams {
			cap := maxPlayerRating
			team.SetRatingCap(&cap)
		}
	}

	lm := &LeagueManager{
		Teams:           teams,
		MaxPlayerRating: maxPlayerRating,
		MaxTeamRating:   maxTeamRating,
		PositionRewards: make(map[int]map[string]int),
		Standings:       make(map[string]*TableStanding),
		PlayerStats:     make(map[string]*PlayerSeasonStats),
		CupEliminated:   make(map[string]bool),
	}

	// Build the Standings Map
	for _, t := range teams {
		lm.Standings[t.ID] = &TableStanding{
			TeamID:   t.ID,
			TeamName: t.Name,
			Form:     make([]string, 0),
		}
	}

	lm.Schedule = buildInterleavedSchedule(teams, roundRobinMultiplier)
	return lm, nil
}

// generateRoundRobin uses the standard circle method to generate a full home-and-away schedule.
// The 'multiplier' determines how many times teams play each other (1 = Single, 2 = Double, 3 = Triple, etc).
func generateRoundRobin(teams []*Team, multiplier int) [][]Fixture {
	if multiplier < 1 {
		multiplier = 1
	}

	n := len(teams)
	halfRounds := n - 1
	rounds := halfRounds * multiplier

	schedule := make([][]Fixture, rounds)
	
	t := make([]*Team, n)
	copy(t, teams)

	// 1. Generate the base Single Round Robin
	for round := 0; round < halfRounds; round++ {
		var roundFixtures []Fixture
		for i := 0; i < n/2; i++ {
			home := t[i]
			away := t[n-1-i]
			// Alternate home/away for the first pair to prevent one team always playing home
			if round%2 == 1 && i == 0 {
				roundFixtures = append(roundFixtures, Fixture{Home: away, Away: home})
			} else {
				roundFixtures = append(roundFixtures, Fixture{Home: home, Away: away})
			}
		}
		schedule[round] = roundFixtures
		
		// Rotate all teams except index 0 (Circle Method)
		last := t[n-1]
		for j := n - 1; j > 1; j-- {
			t[j] = t[j-1]
		}
		t[1] = last
	}
	
	// 2. Generate subsequent rounds by duplicating and flipping Home/Away advantage
	for c := 1; c < multiplier; c++ {
		for round := 0; round < halfRounds; round++ {
			var nextFixtures []Fixture
			for _, f := range schedule[round] {
				// Flip Home/Away on every alternate playthrough (e.g. 2nd time, 4th time)
				if c%2 == 1 {
					nextFixtures = append(nextFixtures, Fixture{Home: f.Away, Away: f.Home})
				} else {
					nextFixtures = append(nextFixtures, Fixture{Home: f.Home, Away: f.Away})
				}
			}
			schedule[(c*halfRounds)+round] = nextFixtures
		}
	}
	
	return schedule
}

// buildInterleavedSchedule takes a base round-robin schedule and inserts Cup rounds evenly throughout the season.
func buildInterleavedSchedule(teams []*Team, multiplier int) []ScheduledRound {
	leagueFixtures := generateRoundRobin(teams, multiplier)
	
	// Calculate how many cup rounds we need for N teams
	c := 0
	if len(teams) > 1 {
		target := 1
		for target < len(teams) {
			target *= 2
			c++
		}
	}
	
	l := len(leagueFixtures)
	var schedule []ScheduledRound
	
	spacing := l / c
	if spacing == 0 {
		spacing = 1
	}
	
	leagueIdx := 0
	cupIdx := 0
	
	for leagueIdx < l {
		schedule = append(schedule, ScheduledRound{Type: MatchLeague, Fixtures: leagueFixtures[leagueIdx]})
		leagueIdx++
		
		if leagueIdx%spacing == 0 && cupIdx < c {
			schedule = append(schedule, ScheduledRound{Type: MatchCup}) // Fixtures are generated dynamically at runtime
			cupIdx++
		}
	}
	for cupIdx < c {
		schedule = append(schedule, ScheduledRound{Type: MatchCup})
		cupIdx++
	}
	
	return schedule
}

// GetNextRound returns the fixtures and match type for the current round and increments the round counter.
// Returns nil if the season is completely finished.
func (lm *LeagueManager) GetNextRound() *ScheduledRound {
	if lm.CurrentRound >= len(lm.Schedule) {
		return nil // Season over
	}
	
	round := &lm.Schedule[lm.CurrentRound]
	lm.CurrentRound++
	
	if round.Type == MatchCup {
		round.Fixtures = lm.generateCupFixtures()
	}
	
	return round
}

// generateCupFixtures dynamically pairs surviving cup teams to play the next knockout round.
func (lm *LeagueManager) generateCupFixtures() []Fixture {
	var active []*Team
	for _, t := range lm.Teams {
		if !lm.CupEliminated[t.ID] {
			active = append(active, t)
		}
	}

	if len(active) <= 1 {
		return nil
	}

	target := 1
	for target*2 <= len(active) {
		target *= 2
	}

	var fixtures []Fixture
	if len(active) == target {
		// Perfect power of 2. Everyone plays.
		for i := 0; i < len(active); i += 2 {
			fixtures = append(fixtures, Fixture{Home: active[i], Away: active[i+1]})
		}
	} else {
		// Prelim round to trim down to the next power of 2
		matchesNeeded := len(active) - target
		for i := 0; i < matchesNeeded*2; i += 2 {
			fixtures = append(fixtures, Fixture{Home: active[i], Away: active[i+1]})
		}
	}

	return fixtures
}

// RecordMatch reads a MatchState payload and permanently records the results.
// For League matches, it updates the standings. For Cup matches, it eliminates the loser.
// For both, it accumulates Player Season Stats.
func (lm *LeagueManager) RecordMatch(state *MatchState) {
	if state.MatchType == MatchLeague {
		// Update Home Team Standings
		homeS := lm.Standings[state.HomeStats.Team.ID]
		if homeS != nil {
			homeS.Played++
			homeS.Wins += state.HomeStats.Wins
			homeS.Draws += state.HomeStats.Draws
			homeS.Losses += state.HomeStats.Losses
			homeS.GoalsFor += state.HomeStats.GoalsFor
			homeS.GoalsAgainst += state.HomeStats.GoalsAgainst
			homeS.GoalDifference = homeS.GoalsFor - homeS.GoalsAgainst
			homeS.Points += (state.HomeStats.Wins * 3) + (state.HomeStats.Draws * 1)
			
			if state.HomeStats.Wins == 1 {
				pushForm(homeS, "W")
			} else if state.HomeStats.Draws == 1 {
				pushForm(homeS, "D")
			} else {
				pushForm(homeS, "L")
			}
		}

		// Update Away Team Standings
		awayS := lm.Standings[state.AwayStats.Team.ID]
		if awayS != nil {
			awayS.Played++
			awayS.Wins += state.AwayStats.Wins
			awayS.Draws += state.AwayStats.Draws
			awayS.Losses += state.AwayStats.Losses
			awayS.GoalsFor += state.AwayStats.GoalsFor
			awayS.GoalsAgainst += state.AwayStats.GoalsAgainst
			awayS.GoalDifference = awayS.GoalsFor - awayS.GoalsAgainst
			awayS.Points += (state.AwayStats.Wins * 3) + (state.AwayStats.Draws * 1)
			
			if state.AwayStats.Wins == 1 {
				pushForm(awayS, "W")
			} else if state.AwayStats.Draws == 1 {
				pushForm(awayS, "D")
			} else {
				pushForm(awayS, "L")
			}
		}
	} else if state.MatchType == MatchCup {
		loser := state.HomeStats.Team.ID
		if state.Winner != nil && *state.Winner == "Home" {
			loser = state.AwayStats.Team.ID
		} else if state.Winner != nil && *state.Winner == "Away" {
			loser = state.HomeStats.Team.ID
		} else if state.HomeStats.GoalsFor > state.AwayStats.GoalsFor {
			loser = state.AwayStats.Team.ID // Fallback check
		}
		lm.CupEliminated[loser] = true
	}

	// Update Player Statistics (Counts for ALL competitions)
	for _, pStats := range state.PlayerStats.Stats {
		if pStats.Appearances == 0 {
			continue
		}
		
		// Ensure the player tracker exists
		ps, exists := lm.PlayerStats[pStats.Player.ID]
		if !exists {
			// Find which team the player belongs to (Home or Away)
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
			lm.PlayerStats[pStats.Player.ID] = ps
		}

		ps.Appearances += pStats.Appearances
		ps.Goals += pStats.Goals
		ps.Assists += pStats.Assists
		ps.CleanSheets += pStats.CleanSheets
		ps.Tackles += pStats.Tackles
	}
}

func belongsTo(playerID string, team *Team) bool {
	for _, p := range team.Players {
		if p.ID == playerID {
			return true
		}
	}
	return false
}

func pushForm(s *TableStanding, res string) {
	s.Form = append(s.Form, res)
	if len(s.Form) > 5 {
		s.Form = s.Form[1:] // Keep only last 5 matches
	}
}

// GetTable returns the fully sorted League Table (Points > GD > GF).
func (lm *LeagueManager) GetTable() []TableStanding {
	var table []TableStanding
	for _, s := range lm.Standings {
		table = append(table, *s)
	}

	sort.Slice(table, func(i, j int) bool {
		if table[i].Points != table[j].Points {
			return table[i].Points > table[j].Points
		}
		if table[i].GoalDifference != table[j].GoalDifference {
			return table[i].GoalDifference > table[j].GoalDifference
		}
		return table[i].GoalsFor > table[j].GoalsFor
	})

	for i := range table {
		table[i].Rank = i + 1
	}

	return table
}

// GetTopScorers returns the top goalscorers sorted by Goals.
func (lm *LeagueManager) GetTopScorers(limit int) []PlayerSeasonStats {
	var list []PlayerSeasonStats
	for _, p := range lm.PlayerStats {
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

// GetTopAssists returns the top playmakers sorted by Assists.
func (lm *LeagueManager) GetTopAssists(limit int) []PlayerSeasonStats {
	var list []PlayerSeasonStats
	for _, p := range lm.PlayerStats {
		if p.Assists > 0 {
			list = append(list, *p)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Assists > list[j].Assists })
	if len(list) > limit {
		return list[:limit]
	}
	return list
}

// GetTopCleanSheets returns the top goalkeepers sorted by Clean Sheets.
func (lm *LeagueManager) GetTopCleanSheets(limit int) []PlayerSeasonStats {
	var list []PlayerSeasonStats
	for _, p := range lm.PlayerStats {
		if p.CleanSheets > 0 {
			list = append(list, *p)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CleanSheets > list[j].CleanSheets })
	if len(list) > limit {
		return list[:limit]
	}
	return list
}

// GetTopDefenders returns the top tacklers in the league.
func (lm *LeagueManager) GetTopDefenders(limit int) []PlayerSeasonStats {
	var list []PlayerSeasonStats
	for _, p := range lm.PlayerStats {
		if p.Tackles > 0 {
			list = append(list, *p)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Tackles > list[j].Tackles })
	if len(list) > limit {
		return list[:limit]
	}
	return list
}

// GetChampionID safely returns the team ID of the 1st place team.
func (lm *LeagueManager) GetChampionID() string {
	table := lm.GetTable()
	if len(table) > 0 {
		return table[0].TeamID
	}
	return ""
}

// GetRunnerUpID safely returns the team ID of the 2nd place team.
func (lm *LeagueManager) GetRunnerUpID() string {
	table := lm.GetTable()
	if len(table) > 1 {
		return table[1].TeamID
	}
	return ""
}
