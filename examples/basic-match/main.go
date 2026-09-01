package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Alliance-Sky/football-sim-engine/engine"
)

// createDummyTeam generates a generic team for demonstration purposes.
func createDummyTeam(id, name, formation, kit string) *engine.Team {
	slots := engine.FormationSlots[formation]
	var players []*engine.Player
	playerIdx := 1

	for pos, count := range slots {
		for i := 0; i < count; i++ {
			pName := fmt.Sprintf("%s Player %d", name, playerIdx)
			p, _ := engine.NewPlayer(
				fmt.Sprintf("%s-p%d", id, playerIdx),
				pName,
				pos,
				engine.FootRight,     // Default right foot
				85.0,                 // 85 Rating
				24,                   // 24 Years old
				100.0,                // 100% Health
				new(engine.Position), 90.0,
			)
			// Explicitly assign natural position to assigned position to prevent OOP penalties
			p.AssignedPosition = pos
			players = append(players, p)
			playerIdx++
		}
	}

	team, err := engine.NewTeam(id, name, formation, kit, players)
	if err != nil {
		log.Fatalf("Failed to create team %s: %v", name, err)
	}
	return team
}

func main() {
	fmt.Println("🚀 Initializing RPS Football Engine...")

	// 1. Create two teams
	home := createDummyTeam("t1", "Red FC", "4-3-3 Attacking", "#ff0000")
	away := createDummyTeam("t2", "Blue FC", "4-2-3-1", "#0000ff")

	// 2. Play the match using the QuickPlay facade
	fmt.Println("⚽ Simulating match...")
	state, err := engine.QuickPlay(engine.MatchLeague, home, away, true)
	if err != nil {
		log.Fatalf("Match simulation failed: %v", err)
	}

	// 3. Output the final scoreline
	fmt.Printf("\n--- FULL TIME ---\n")
	fmt.Printf("%s %d : %d %s\n\n", home.Name, state.HomeStats.MatchGoalsFor, state.AwayStats.MatchGoalsFor, away.Name)

	// 4. Output the delta JSON payload
	fmt.Println("📦 Match State Delta Payload (JSON):")
	payload, _ := json.MarshalIndent(state, "", "  ")
	fmt.Println(string(payload))
}
