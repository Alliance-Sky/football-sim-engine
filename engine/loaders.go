package engine

import (
	"encoding/json"
	"fmt"
)

// LoadPlayersFromJSON takes a raw JSON array of player data and returns a slice of validated *Player objects.
// This is useful for hydrating players directly from a database payload.
func LoadPlayersFromJSON(data []byte) ([]*Player, error) {
	var rawPlayers []Player
	if err := json.Unmarshal(data, &rawPlayers); err != nil {
		return nil, err
	}

	var validPlayers []*Player
	for _, rp := range rawPlayers {
		var assigned *Position
		if rp.AssignedPosition != "" {
			assigned = &rp.AssignedPosition
		}

		p, err := NewPlayer(rp.ID, rp.Name, rp.NaturalPosition, rp.Foot, rp.Rating, rp.Age, rp.Health, assigned)
		if err != nil {
			return nil, fmt.Errorf("failed to validate player %s (%s): %v", rp.Name, rp.ID, err)
		}
		validPlayers = append(validPlayers, p)
	}
	return validPlayers, nil
}

// LoadTeamFromJSON is a powerful helper that parses a JSON array of players and constructs a fully validated Team.
func LoadTeamFromJSON(id, name, formation, kitColor string, playerData []byte) (*Team, error) {
	players, err := LoadPlayersFromJSON(playerData)
	if err != nil {
		return nil, err
	}
	return NewTeam(id, name, formation, kitColor, players)
}
