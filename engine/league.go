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

// LeagueManager tracks scheduling, team standings, and player statistics for a full season.
type LeagueManager struct {
	Teams        []*Team
	Schedule     [][]Fixture
	CurrentRound int

	Standings   map[string]*TableStanding
	PlayerStats map[string]*PlayerSeasonStats
}

// NewLeagueManager initializes a new league, builds the standings map, and generates the round-robin schedule.
func NewLeagueManager(teams []*Team) (*LeagueManager, error) {
	if len(teams) < 2 {
		return nil, fmt.Errorf("league requires at least 2 teams")
	}
	if len(teams)%2 != 0 {
		return nil, fmt.Errorf("league requires an even number of teams for scheduling")
	}

	lm := &LeagueManager{
		Teams:       teams,
		Standings:   make(map[string]*TableStanding),
		PlayerStats: make(map[string]*PlayerSeasonStats),
	}

	for _, t := range teams {
		lm.Standings[t.ID] = &TableStanding{
			TeamID:   t.ID,
			TeamName: t.Name,
			Form:     make([]string, 0),
		}
	}

	lm.Schedule = generateRoundRobin(teams)
	return lm, nil
}

// generateRoundRobin uses the standard circle method to generate a full home-and-away schedule.
func generateRoundRobin(teams []*Team) [][]Fixture {
	n := len(teams)
	halfRounds := n - 1
	rounds := halfRounds * 2
	schedule := make([][]Fixture, rounds)
	
	t := make([]*Team, n)
	copy(t, teams)

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
	
	// Generate the reverse fixtures for the second half of the season
	for round := 0; round < halfRounds; round++ {
		var reverseFixtures []Fixture
		for _, f := range schedule[round] {
			reverseFixtures = append(reverseFixtures, Fixture{Home: f.Away, Away: f.Home})
		}
		schedule[halfRounds+round] = reverseFixtures
	}
	
	return schedule
}

// GetNextRound returns the fixtures for the current round and increments the round counter.
// Returns nil if the season is completely finished.
func (lm *LeagueManager) GetNextRound() []Fixture {
	if lm.CurrentRound >= len(lm.Schedule) {
		return nil // Season over
	}
	fixtures := lm.Schedule[lm.CurrentRound]
	lm.CurrentRound++
	return fixtures
}

// RecordMatch reads a MatchState payload and permanently records the results to the league tables.
func (lm *LeagueManager) RecordMatch(state *MatchState) {
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

	// Update Player Statistics
	for _, pStats := range state.PlayerStats.Stats {
		if pStats.Appearances == 0 {
			continue
		}
		
		// Ensure the player tracker exists
		ps, exists := lm.PlayerStats[pStats.Player.ID]
		if !exists {
			// Find which team the player belongs to (Home or Away)
			teamID, teamName := "", ""
			if homeS != nil && belongsTo(pStats.Player.ID, state.HomeStats.Team) {
				teamID, teamName = state.HomeStats.Team.ID, state.HomeStats.Team.Name
			} else if awayS != nil && belongsTo(pStats.Player.ID, state.AwayStats.Team) {
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

// GetTopDefenders returns the top defensive players sorted by Tackles won.
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
