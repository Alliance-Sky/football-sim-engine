// Package engine provides core simulation logic and statistics tracking for the RPS Football Engine.
package engine

// PlayerMatchStats holds the performance and health metrics for a single player
// in a single match. It tracks stats like goals, assists, clean sheets, and tackles.
type PlayerMatchStats struct {
	Player *Player `json:"player"`

	MatchGoals       int     `json:"matchGoals"`
	MatchAssists     int     `json:"matchAssists"`
	MatchCleanSheets int     `json:"matchCleanSheets"`
	MatchTackles     int     `json:"matchTackles"`
	MatchHealthLost  float64 `json:"matchHealthLost"`

	PostMatchMatches     int     `json:"postMatchMatches"`
	PostMatchGoals       int     `json:"postMatchGoals"`
	PostMatchAssists     int     `json:"postMatchAssists"`
	PostMatchCleanSheets int     `json:"postMatchCleanSheets"`
	PostMatchTackles     int     `json:"postMatchTackles"`
	PostMatchHealth      float64 `json:"postMatchHealth"`
}

// StatsTracker manages the collection of stats for all players involved in a match.
// It maps players to their match-specific PlayerMatchStats structure.
type StatsTracker struct {
	Stats map[string]*PlayerMatchStats `json:"stats"` // Map of player IDs to their stats
}

// NewStatsTracker creates and initializes a new StatsTracker.
func NewStatsTracker() *StatsTracker {
	return &StatsTracker{
		Stats: make(map[string]*PlayerMatchStats),
	}
}

// GetOrCreate retrieves the PlayerMatchStats for a given player,
// or creates a new one and adds it to the tracker if it doesn't exist yet.
func (st *StatsTracker) GetOrCreate(p *Player) *PlayerMatchStats {
	if entry, exists := st.Stats[p.ID]; exists {
		return entry
	}
	entry := &PlayerMatchStats{Player: p}

	// Initialize with the player's current cumulative totals
	entry.PostMatchMatches = p.Matches
	entry.PostMatchGoals = p.Goals
	entry.PostMatchAssists = p.Assists
	entry.PostMatchCleanSheets = p.CleanSheets
	entry.PostMatchTackles = p.Tackles
	entry.PostMatchHealth = p.Health

	st.Stats[p.ID] = entry
	return entry
}

// RecordGoal increments the goal count for the specified player.
func (st *StatsTracker) RecordGoal(scorer *Player) {
	if scorer == nil {
		return
	}
	scorer.Goals++
	entry := st.GetOrCreate(scorer)
	entry.MatchGoals++
	entry.PostMatchGoals = scorer.Goals
}

// RecordAssist increments the assist count for the specified player.
func (st *StatsTracker) RecordAssist(assister *Player) {
	if assister == nil {
		return
	}
	assister.Assists++
	entry := st.GetOrCreate(assister)
	entry.MatchAssists++
	entry.PostMatchAssists = assister.Assists
}

// RecordCleanSheets awards a clean sheet to the goalkeepers of any team
// that did not concede a goal during the match.
func (st *StatsTracker) RecordCleanSheets(home, away *Team, homeScore, awayScore int) {
	if awayScore == 0 {
		for _, gk := range home.Goalkeepers() {
			gk.CleanSheets++
			entry := st.GetOrCreate(gk)
			entry.MatchCleanSheets++
			entry.PostMatchCleanSheets = gk.CleanSheets
		}
	}
	if homeScore == 0 {
		for _, gk := range away.Goalkeepers() {
			gk.CleanSheets++
			entry := st.GetOrCreate(gk)
			entry.MatchCleanSheets++
			entry.PostMatchCleanSheets = gk.CleanSheets
		}
	}
}

// RecordMatchPlayed marks the player as having appeared in the match.
func (st *StatsTracker) RecordMatchPlayed(p *Player) {
	if p == nil {
		return
	}
	// We only want to increment this once per match, not every tick.
	// We will track matches when they are first initialized in a match
	// or handled elsewhere. Let's assume RecordMatchPlayed is only called once
	// per match for starters in engine/matches.go
	p.Matches++
	st.GetOrCreate(p).PostMatchMatches = p.Matches
}

// RecordTackle increments the tackle count for the specified player.
func (st *StatsTracker) RecordTackle(p *Player) {
	if p == nil {
		return
	}
	p.Tackles++
	entry := st.GetOrCreate(p)
	entry.MatchTackles++
	entry.PostMatchTackles = p.Tackles
}
