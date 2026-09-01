// Package engine provides core simulation logic and statistics tracking for the RPS Football Engine.
package engine

// PlayerMatchStats holds the performance and health metrics for a single player
// in a single match. It tracks stats like goals, assists, clean sheets, and tackles.
type PlayerMatchStats struct {
	Player      *Player // The player these stats belong to
	Appearances int     // 1 if the player started the match
	Goals       int     // Goals scored by the player
	Assists     int     // Assists made by the player
	CleanSheets int     // Earned if the player is a GK and their team conceded 0 goals
	Tackles     int     // Successful tackles made by the player
	HealthLost  float64 // Amount of health (fatigue) lost during the match
}

// StatsTracker manages the collection of stats for all players involved in a match.
// It maps players to their match-specific PlayerMatchStats structure.
type StatsTracker struct {
	Stats map[*Player]*PlayerMatchStats // Map of players to their stats
}

// NewStatsTracker creates and initializes a new StatsTracker.
func NewStatsTracker() *StatsTracker {
	return &StatsTracker{
		Stats: make(map[*Player]*PlayerMatchStats),
	}
}

// GetOrCreate retrieves the PlayerMatchStats for a given player,
// or creates a new one and adds it to the tracker if it doesn't exist yet.
func (st *StatsTracker) GetOrCreate(p *Player) *PlayerMatchStats {
	if entry, exists := st.Stats[p]; exists {
		return entry
	}
	entry := &PlayerMatchStats{Player: p}
	st.Stats[p] = entry
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
