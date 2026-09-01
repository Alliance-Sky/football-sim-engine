package engine

import (
	"fmt"
	"math/rand"
	"time"
)

// GroupTournamentManager orchestrates a World Cup style tournament:
// A Group Stage (mini round-robins) followed by a Knockout Stage.
type GroupTournamentManager struct {
	ID              string
	Name            string
	MaxParticipants int // e.g. 32
	GroupSize       int // Usually 4

	Participants []*Team
	Groups       map[string]*LeagueManager // "A", "B", "C"...
	Knockout     *TournamentManager

	Stage string // "Lobby", "Groups", "Knockout", "Finished"

	rng *rand.Rand
}

// NewGroupTournamentManager initializes a Group + Knockout tournament.
func NewGroupTournamentManager(id, name string, maxParticipants, groupSize int) (*GroupTournamentManager, error) {
	if maxParticipants%groupSize != 0 {
		return nil, fmt.Errorf("MaxParticipants must be perfectly divisible by GroupSize")
	}

	return &GroupTournamentManager{
		ID:              id,
		Name:            name,
		MaxParticipants: maxParticipants,
		GroupSize:       groupSize,
		Participants:    make([]*Team, 0),
		Groups:          make(map[string]*LeagueManager),
		Stage:           "Lobby",
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// Start draws the groups and initializes the group stage leagues.
func (gtm *GroupTournamentManager) Start() error {
	if len(gtm.Participants) < gtm.MaxParticipants {
		gtm.Participants = FillWithBots(gtm.Participants, gtm.MaxParticipants, 65.0)
	}

	// Shuffle teams for the draw
	gtm.rng.Shuffle(len(gtm.Participants), func(i, j int) {
		gtm.Participants[i], gtm.Participants[j] = gtm.Participants[j], gtm.Participants[i]
	})

	numGroups := gtm.MaxParticipants / gtm.GroupSize

	// Create Groups (A, B, C...)
	for i := 0; i < numGroups; i++ {
		groupName := string(rune('A' + i))

		startIndex := i * gtm.GroupSize
		endIndex := startIndex + gtm.GroupSize
		groupTeams := gtm.Participants[startIndex:endIndex]

		// Each group is literally just a mini LeagueManager (Single Round Robin!)
		groupLeague, _ := NewLeagueManager(groupTeams, 0, 0, 1)
		gtm.Groups[groupName] = groupLeague
	}

	gtm.Stage = "Groups"
	return nil
}

// TransitionToKnockouts pulls the Top 2 from every group and builds the knockout bracket.
func (gtm *GroupTournamentManager) TransitionToKnockouts() error {
	if gtm.Stage != "Groups" {
		return fmt.Errorf("must be in Groups stage")
	}

	var advancingTeams []*Team

	// Iterate through groups A, B, C...
	numGroups := gtm.MaxParticipants / gtm.GroupSize
	for i := 0; i < numGroups; i++ {
		groupName := string(rune('A' + i))
		league := gtm.Groups[groupName]

		table := league.GetTable()
		if len(table) >= 2 {
			// Find the actual *Team structs by matching the ID
			for _, t := range league.Teams {
				if t.ID == table[0].TeamID || t.ID == table[1].TeamID {
					advancingTeams = append(advancingTeams, t)
				}
			}
		}
	}

	// Create the Knockout Bracket!
	knockout, _ := NewTournamentManager(gtm.ID+"-KO", gtm.Name+" Finals", len(advancingTeams), 0, 0, false)

	// Manually inject the active teams
	knockout.Participants = advancingTeams
	knockout.ActiveTeams = make([]*Team, len(advancingTeams))
	copy(knockout.ActiveTeams, advancingTeams)

	gtm.Knockout = knockout
	gtm.Stage = "Knockout"
	return nil
}
