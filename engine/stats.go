// Package engine provides core simulation logic and statistics tracking for the RPS Football Engine.
package engine

// PlayerMatchStats holds the performance and health metrics for a single player
// in a single match. It tracks stats like goals, assists, clean sheets, and tackles.
type PlayerMatchStats struct {
	Player      *Player `json:"player"`
	Appearances int     `json:"appearances"`
	Goals       int     `json:"goals"`
	Assists     int     `json:"assists"`
	CleanSheets int     `json:"cleanSheets"`
	Tackles     int     `json:"tackles"`
	HealthLost  float64 `json:"healthLost"`
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
	st.Stats[p.ID] = entry
	return entry
}

// RecordGoal increments the goal count for the specified player.
func (st *StatsTracker) RecordGoal(scorer *Player) {
	if scorer == nil {
		return
	}
	st.GetOrCreate(scorer).Goals++
}

// RecordAssist increments the assist count for the specified player.
func (st *StatsTracker) RecordAssist(assister *Player) {
	if assister == nil {
		return
	}
	st.GetOrCreate(assister).Assists++
}

// RecordCleanSheets awards a clean sheet to the goalkeepers of any team
// that did not concede a goal during the match.
func (st *StatsTracker) RecordCleanSheets(home, away *Team, homeScore, awayScore int) {
	if awayScore == 0 {
		for _, gk := range home.Goalkeepers() {
			st.GetOrCreate(gk).CleanSheets++
		}
	}
	if homeScore == 0 {
		for _, gk := range away.Goalkeepers() {
			st.GetOrCreate(gk).CleanSheets++
		}
	}
}

// RecordAppearance marks the player as having appeared in the match.
func (st *StatsTracker) RecordAppearance(p *Player) {
	if p == nil {
		return
	}
	st.GetOrCreate(p).Appearances = 1
}

// RecordTackle increments the tackle count for the specified player.
func (st *StatsTracker) RecordTackle(p *Player) {
	if p == nil {
		return
	}
	st.GetOrCreate(p).Tackles++
}
