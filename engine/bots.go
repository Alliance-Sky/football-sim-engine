package engine

import (
	"fmt"
	"math/rand"
)

// BotCities is a global list of cities used for generating realistic bot names.
var BotCities = []string{
	"London", "Berlin", "Paris", "Madrid", "Rome",
	"Tokyo", "New York", "Sydney", "Rio", "Cairo",
	"Seoul", "Mexico City", "Mumbai", "Jakarta", "Lagos",
	"Toronto", "Dubai", "Istanbul", "Moscow", "Bangkok",
	"Beijing", "Amsterdam", "Dublin", "Vienna", "Lisbon",
}

// FillWithBots takes an existing slice of teams and appends perfectly scaled AI bots
// until the slice reaches the required maxCapacity.
func FillWithBots(teams []*Team, maxCapacity int, targetRating float64) []*Team {
	needed := maxCapacity - len(teams)
	if needed <= 0 {
		return teams
	}

	for i := 0; i < needed; i++ {
		city := BotCities[i%len(BotCities)]
		botName := fmt.Sprintf("%s Bot FC", city)
		botID := fmt.Sprintf("bot-gen-%d", len(teams)+1) // Ensure unique ID

		botTeam := GenerateBotTeam(botID, botName, targetRating)
		teams = append(teams, botTeam)
	}

	return teams
}

// GenerateBotTeam is a helper that generates a perfectly valid AI team at a specific target rating.
func GenerateBotTeam(id, name string, targetRating float64) *Team {
	formation := "4-4-2 Flat"
	slots := FormationSlots[formation]
	var players []*Player
	playerIdx := 1

	for pos, count := range slots {
		for i := 0; i < count; i++ {
			// Add slight randomness (-2 to +2) to bot ratings
			fuzz := (rand.Float64() * 4) - 2
			finalRating := targetRating + fuzz

			p, _ := NewPlayer(
				fmt.Sprintf("%s-p%d", id, playerIdx),
				fmt.Sprintf("%s Player %d", name, playerIdx),
				pos, FootRight, finalRating, 25, 100.0, nil,
			)
			p.AssignedPosition = pos
			players = append(players, p)
			playerIdx++
		}
	}
	t, _ := NewTeam(id, name, formation, "#cccccc", players)
	return t
}
