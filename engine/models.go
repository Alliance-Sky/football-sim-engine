// Package engine provides core simulation logic and models for the RPS Football Engine.
package engine

import (
	"fmt"
	"strings"
)

// Position represents a player's position on the pitch.
type Position string

const (
	PosGK Position = "GK"
	PosLB Position = "LB"
	PosCB Position = "CB"
	PosRB Position = "RB"
	PosLM Position = "LM"
	PosDM Position = "DM"
	PosCM Position = "CM"
	PosRM Position = "RM"
	PosLW Position = "LW"
	PosST Position = "ST"
	PosRW Position = "RW"
)

// GetPositions returns a slice containing all valid on-pitch positions.
func GetPositions() []Position {
	return []Position{
		PosGK, PosLB, PosCB, PosRB, PosDM, PosCM, PosLM, PosRM, PosLW, PosST, PosRW,
	}
}

// Helper methods to categorize positions.
func (p Position) IsGoalkeeper() bool   { return p == PosGK }
func (p Position) IsCenterBack() bool   { return p == PosCB }
func (p Position) IsFullBack() bool     { return p == PosLB || p == PosRB }
func (p Position) IsDefensiveMid() bool { return p == PosDM }
func (p Position) IsCentralMid() bool   { return p == PosCM }
func (p Position) IsWideMid() bool      { return p == PosLM || p == PosRM }
func (p Position) IsWinger() bool       { return p == PosLW || p == PosRW }
func (p Position) IsStriker() bool      { return p == PosST }
func (p Position) IsDefender() bool     { return p == PosLB || p == PosCB || p == PosRB }
func (p Position) IsMidfielder() bool   { return p == PosLM || p == PosDM || p == PosCM || p == PosRM }
func (p Position) IsForward() bool      { return p == PosLW || p == PosST || p == PosRW }

// MatchType defines whether a match is a league match or a cup knockout match.
type MatchType string

const (
	MatchLeague MatchType = "League Match"
	MatchCup    MatchType = "Cup Knockout Match"
)

// PitchZone divides the pitch into 5 horizontal zones to track ball location.
type PitchZone int

const (
	ZoneHomeBox  PitchZone = 1 // Inside the home team's penalty box
	ZoneHomeHalf PitchZone = 2 // In the home team's defensive half
	ZoneMidfield PitchZone = 3 // Around the center circle
	ZoneAwayHalf PitchZone = 4 // In the away team's defensive half
	ZoneAwayBox  PitchZone = 5 // Inside the away team's penalty box
)

// Player represents a footballer with attributes and status.
type Player struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	NaturalPosition  Position  `json:"naturalPosition"`
	Foot             Foot      `json:"foot"`
	Rating           float64   `json:"rating"`
	Age              int       `json:"age"`
	Health           float64   `json:"health"`
	AssignedPosition Position  `json:"assignedPosition"`
	RatingCap        *float64  `json:"ratingCap,omitempty"`
}

// NewPlayer creates and validates a new Player instance.
func NewPlayer(id, name string, naturalPos Position, foot Foot, rating float64, age int, health float64, assignedPos *Position) (*Player, error) {
	if foot != FootLeft && foot != FootRight && foot != FootBoth {
		return nil, fmt.Errorf("player '%s' foot '%s' is invalid; must be 'Left', 'Right', or 'Both'", name, foot)
	}
	if rating < MinRating || rating > MaxRating {
		return nil, fmt.Errorf("player '%s' rating %.1f is invalid; must be between %.0f and %.0f", name, rating, MinRating, MaxRating)
	}
	if age < MinAge || age > MaxAge {
		return nil, fmt.Errorf("player '%s' age %d is invalid; must be between %d and %d", name, age, MinAge, MaxAge)
	}
	if health < MinHealth || health > MaxHealth {
		return nil, fmt.Errorf("player '%s' health %.1f is invalid; must be between %.0f and %.0f", name, health, MinHealth, MaxHealth)
	}

	actualAssigned := naturalPos
	if assignedPos != nil {
		actualAssigned = *assignedPos
	}

	return &Player{
		ID:               id,
		Name:             name,
		NaturalPosition:  naturalPos,
		Foot:             foot,
		Rating:           rating,
		Age:              age,
		Health:           health,
		AssignedPosition: actualAssigned,
		RatingCap:        nil,
	}, nil
}

// IsOutOfPosition returns true if the player is not playing in their natural position.
func (p *Player) IsOutOfPosition() bool {
	return p.NaturalPosition != p.AssignedPosition
}

// OOPPenalty calculates the penalty percentage to apply to the player's rating due to being out of position.
func (p *Player) OOPPenalty() float64 {
	return CalculateOOPPenalty(p.NaturalPosition, p.AssignedPosition, p.Foot)
}

// CappedRating returns the player's base rating, bounded by the RatingCap if one is set.
func (p *Player) CappedRating() float64 {
	if p.RatingCap != nil {
		return max(MinRating, min(p.Rating, *p.RatingCap))
	}
	return max(MinRating, p.Rating)
}

// HealthDeficit calculates how much health the player has lost compared to MaxHealth.
func (p *Player) HealthDeficit() float64 {
	return max(0.0, MaxHealth-p.Health)
}

// PostHealthRating reduces the capped rating by the health deficit. (1 health lost = 1 rating point lost).
func (p *Player) PostHealthRating() float64 {
	reduced := p.CappedRating() - p.HealthDeficit()
	return max(MinRating, min(MaxRating, reduced))
}

// EffectiveRating computes the final rating used during match simulations, factoring in health and out-of-position penalties.
func (p *Player) EffectiveRating() float64 {
	// Apply the OOP percentage penalty to the health-adjusted rating.
	penalized := p.PostHealthRating() * (1.0 - p.OOPPenalty())
	return max(MinRating, min(MaxRating, penalized))
}

// Team represents a group of 11 players configured in a specific formation.
type Team struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Formation string    `json:"formation"`
	KitColor  string    `json:"kitColor"`
	Players   []*Player `json:"players"`
}

// NewTeam creates a Team and validates that its lineup and mandatory fields are valid.
func NewTeam(id, name, formation, kitColor string, players []*Player) (*Team, error) {
	if strings.TrimSpace(kitColor) == "" {
		return nil, fmt.Errorf("team '%s' requires a mandatory non-empty kitColor", name)
	}
	t := &Team{
		ID:        id,
		Name:      name,
		Formation: formation,
		KitColor:  kitColor,
		Players:   players,
	}
	if err := t.ValidateLineup(); err != nil {
		return nil, err
	}
	return t, nil
}

// SetRatingCap applies a global rating cap to all players on the team.
func (t *Team) SetRatingCap(cap *float64) {
	for _, p := range t.Players {
		p.RatingCap = cap
	}
}

// Style returns the tactical style (Rock, Paper, Scissors) of the team's formation.
func (t *Team) Style() string {
	return Formations[t.Formation]
}

// ValidateLineup checks if the team has exactly 11 players and fills the required positional slots for the formation.
func (t *Team) ValidateLineup() error {
	reqSlots, exists := FormationSlots[t.Formation]
	if !exists {
		return fmt.Errorf("unknown formation '%s'", t.Formation)
	}
	if len(t.Players) != 11 {
		return fmt.Errorf("team '%s' must have exactly 11 players, found %d", t.Name, len(t.Players))
	}

	counts := make(map[Position]int)
	for _, p := range t.Players {
		counts[p.AssignedPosition]++
	}

	for pos, req := range reqSlots {
		if counts[pos] != req {
			return fmt.Errorf("lineup mismatch for team '%s' in %s: expected %d for %s, found %d",
				t.Name, t.Formation, req, pos, counts[pos])
		}
	}
	return nil
}

// Positional filtering methods to get subsets of the team's players based on assigned positions.
func (t *Team) Goalkeepers() []*Player {
	return t.filterPlayers(func(p *Player) bool { return p.AssignedPosition.IsGoalkeeper() })
}
func (t *Team) CenterBacks() []*Player {
	return t.filterPlayers(func(p *Player) bool { return p.AssignedPosition.IsCenterBack() })
}
func (t *Team) FullBacks() []*Player {
	return t.filterPlayers(func(p *Player) bool { return p.AssignedPosition.IsFullBack() })
}
func (t *Team) DefensiveMids() []*Player {
	return t.filterPlayers(func(p *Player) bool { return p.AssignedPosition.IsDefensiveMid() })
}
func (t *Team) CentralMids() []*Player {
	return t.filterPlayers(func(p *Player) bool { return p.AssignedPosition.IsCentralMid() })
}
func (t *Team) WideMids() []*Player {
	return t.filterPlayers(func(p *Player) bool { return p.AssignedPosition.IsWideMid() })
}
func (t *Team) Wingers() []*Player {
	return t.filterPlayers(func(p *Player) bool { return p.AssignedPosition.IsWinger() })
}
func (t *Team) Strikers() []*Player {
	return t.filterPlayers(func(p *Player) bool { return p.AssignedPosition.IsStriker() })
}
func (t *Team) Defenders() []*Player {
	return t.filterPlayers(func(p *Player) bool { return p.AssignedPosition.IsDefender() })
}
func (t *Team) Midfielders() []*Player {
	return t.filterPlayers(func(p *Player) bool { return p.AssignedPosition.IsMidfielder() })
}
func (t *Team) Forwards() []*Player {
	return t.filterPlayers(func(p *Player) bool { return p.AssignedPosition.IsForward() })
}

// filterPlayers applies a predicate function to filter the list of team players.
func (t *Team) filterPlayers(predicate func(*Player) bool) []*Player {
	var list []*Player
	for _, p := range t.Players {
		if predicate(p) {
			list = append(list, p)
		}
	}
	return list
}

// avg calculates the mean effective rating of a group of players. Returns fallback if the group is empty.
func (t *Team) avg(group []*Player, fallback float64) float64 {
	if len(group) == 0 {
		return max(MinRating, fallback)
	}
	var total float64
	for _, p := range group {
		total += p.EffectiveRating()
	}
	return max(MinRating, total/float64(len(group)))
}

// BaseOverallRating averages the unadjusted ratings of all 11 players.
func (t *Team) BaseOverallRating() float64 {
	var total float64
	for _, p := range t.Players {
		total += p.Rating
	}
	return total / float64(len(t.Players))
}

// CappedOverallRating averages the capped ratings of all 11 players.
func (t *Team) CappedOverallRating() float64 {
	var total float64
	for _, p := range t.Players {
		total += p.CappedRating()
	}
	return total / float64(len(t.Players))
}

// PostHealthOverallRating averages the health-adjusted ratings of all 11 players.
func (t *Team) PostHealthOverallRating() float64 {
	var total float64
	for _, p := range t.Players {
		total += p.PostHealthRating()
	}
	return total / float64(len(t.Players))
}

// OverallRating averages the final effective ratings (health + OOP factored in) of all 11 players.
func (t *Team) OverallRating() float64 {
	var total float64
	for _, p := range t.Players {
		total += p.EffectiveRating()
	}
	return max(MinRating, total/float64(len(t.Players)))
}

// Sectional and Positional Average Ratings.
func (t *Team) GKRating() float64 {
	gks := t.Goalkeepers()
	if len(gks) > 0 {
		return max(MinRating, gks[0].EffectiveRating())
	}
	return MinRating
}
func (t *Team) CBRating() float64      { return t.avg(t.CenterBacks(), t.OverallRating()) }
func (t *Team) FBRating() float64      { return t.avg(t.FullBacks(), t.CBRating()) }
func (t *Team) DMRating() float64      { return t.avg(t.DefensiveMids(), t.OverallRating()) }
func (t *Team) CMRating() float64      { return t.avg(t.CentralMids(), t.OverallRating()) }
func (t *Team) WideMidRating() float64 { return t.avg(t.WideMids(), t.CMRating()) }
func (t *Team) WingerRating() float64  { return t.avg(t.Wingers(), t.STRating()) }
func (t *Team) STRating() float64      { return t.avg(t.Strikers(), t.OverallRating()) }
func (t *Team) DefRating() float64     { return t.avg(t.Defenders(), t.OverallRating()) }
func (t *Team) MidRating() float64     { return t.avg(t.Midfielders(), t.OverallRating()) }
func (t *Team) FwdRating() float64     { return t.avg(t.Forwards(), t.OverallRating()) }

// DefenseUnitRating calculates the combined defensive strength of the team.
// It heavily weights center backs, but factors in defensive mids and fullbacks depending on the formation.
func (t *Team) DefenseUnitRating() float64 {
	dmVal := t.CBRating()
	if len(t.DefensiveMids()) > 0 {
		dmVal = t.DMRating()
	}
	fbVal := t.CBRating()
	if len(t.FullBacks()) > 0 {
		fbVal = t.FBRating()
	}

	cbWeight := 0.30
	dmWeight := 0.12
	fbWeight := 0.08

	// Formations with 3 center backs heavily rely on the center backs for defense.
	if len(t.CenterBacks()) > 2 {
		cbWeight = 0.35
		dmWeight = 0.10
		fbWeight = 0.05
	}

	// The goalkeeper is the most important factor in the defense unit (50% weight).
	val := (t.GKRating() * 0.50) + (t.CBRating() * cbWeight) + (dmVal * dmWeight) + (fbVal * fbWeight)
	return max(MinRating, val)
}

// ProgressionRating measures a team's ability to move the ball up the pitch.
// Relies heavily on central and wide midfielders, with support from fullbacks and wingers.
func (t *Team) ProgressionRating() float64 {
	midCreators := append(t.CentralMids(), t.WideMids()...)
	rMid := t.avg(midCreators, t.OverallRating())

	flankProgressors := append(t.FullBacks(), t.Wingers()...)
	rFlank := t.avg(flankProgressors, rMid)

	rDM := rMid
	if len(t.DefensiveMids()) > 0 {
		rDM = t.DMRating()
	}
	rCB := t.CBRating()

	// Midfielders carry 45% of the progression weight, flanks 30%.
	val := (rMid * 0.45) + (rFlank * 0.30) + (rDM * 0.15) + (rCB * 0.10)
	return max(MinRating, val)
}

// CreationThreatRating measures the team's ability to create chances in the final third.
// Averaged across Wingers, Central Mids, and Wide Mids.
func (t *Team) CreationThreatRating() float64 {
	creators := append(append(t.Wingers(), t.CentralMids()...), t.WideMids()...)
	rating := t.avg(creators, t.OverallRating())
	return max(MinRating, rating)
}

// DefensiveShieldRating measures the shielding power of players in front of the center backs.
// Combines Defensive Midfielders and Fullbacks.
func (t *Team) DefensiveShieldRating() float64 {
	shields := append(t.DefensiveMids(), t.FullBacks()...)
	return t.avg(shields, t.DefRating())
}

// ClubMatchStats tracks a specific team's stats during a match simulation.
type ClubMatchStats struct {
	Team            *Team   `json:"team"`
	TeamRating      float64 `json:"teamRating"`
	Matches         int     `json:"matches"`
	Wins            int     `json:"wins"`
	Draws           int     `json:"draws"`
	Losses          int     `json:"losses"`
	GoalsFor        int     `json:"goalsFor"`
	GoalsAgainst    int     `json:"goalsAgainst"`
	XG              float64 `json:"xg"`
	Shots           int     `json:"shots"`
	SOT             int     `json:"sot"`
	Saves           int     `json:"saves"`
	PossessionTicks int     `json:"possessionTicks"`
	Assists         int     `json:"assists"`
	Tackles         int     `json:"tackles"`
}

// MatchLog represents a single play-by-play event with its corresponding minute.
type MatchLog struct {
	Minute  int    `json:"minute"`
	Message string `json:"message"`
}

// MatchState holds the ongoing state of a match being simulated.
type MatchState struct {
	MatchType          MatchType       `json:"matchType"`
	Minute             int             `json:"minute"`
	HomeStats          *ClubMatchStats `json:"homeStats"`
	AwayStats          *ClubMatchStats `json:"awayStats"`
	PlayerStats        *StatsTracker   `json:"playerStats"`
	PossessionTeam     string          `json:"possessionTeam"`
	BallZone           PitchZone       `json:"ballZone"`
	WentToExtraTime    bool            `json:"wentToExtraTime"`
	WentToPenalties    bool            `json:"wentToPenalties"`
	PenaltyShootoutLog []string        `json:"penaltyShootoutLog"`
	Winner             *string         `json:"winner,omitempty"`
	Commentary         []MatchLog      `json:"commentary"`
}

// NewMatchState initializes a new match simulation state.
func NewMatchState(mType MatchType, home, away *Team) *MatchState {
	return &MatchState{
		MatchType:          mType,
		HomeStats:          &ClubMatchStats{Team: home, TeamRating: home.OverallRating(), Matches: 1},
		AwayStats:          &ClubMatchStats{Team: away, TeamRating: away.OverallRating(), Matches: 1},
		BallZone:           ZoneMidfield,
		PossessionTeam:     "Home",
		PlayerStats:        NewStatsTracker(),
		PenaltyShootoutLog: make([]string, 0),
		Commentary:         make([]MatchLog, 0),
	}
}

// Log adds a message to the match commentary timeline.
func (s *MatchState) Log(msg string) {
	s.Commentary = append(s.Commentary, MatchLog{
		Minute:  s.Minute,
		Message: msg,
	})
}
