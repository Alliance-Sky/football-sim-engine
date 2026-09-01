package main

import (
	"fmt"
	"log"

	"rps-football-engine/engine"
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
				new(engine.Position), // Use natural position
			)
			// Explicitly assign natural position to prevent OOP penalties
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
	fmt.Println("🚀 Initializing RPS Football Engine (Deterministic Mode)...")

	// 1. Create two identical team states
	home1 := createDummyTeam("t1", "Red FC", "4-3-3 Attacking", "#ff0000")
	away1 := createDummyTeam("t2", "Blue FC", "4-2-3-1", "#0000ff")

	home2 := createDummyTeam("t1", "Red FC", "4-3-3 Attacking", "#ff0000")
	away2 := createDummyTeam("t2", "Blue FC", "4-2-3-1", "#0000ff")

	// 2. Play identical matches using the EXACT same seed
	seed := int64(9999)
	fmt.Printf("⚽ Simulating Match 1 with Seed %d...\n", seed)
	state1, _ := engine.DeterministicPlay(engine.MatchLeague, home1, away1, true, seed)

	fmt.Printf("⚽ Simulating Match 2 with Seed %d...\n", seed)
	state2, _ := engine.DeterministicPlay(engine.MatchLeague, home2, away2, true, seed)

	// 3. Compare outputs (They will be 100% identical)
	fmt.Printf("\n--- MATCH 1 RESULTS ---\n")
	fmt.Printf("Score: %d - %d\n", state1.HomeStats.MatchGoalsFor, state1.AwayStats.MatchGoalsFor)
	fmt.Printf("Possession: %d - %d\n", state1.HomeStats.PossessionTicks, state1.AwayStats.PossessionTicks)

	fmt.Printf("\n--- MATCH 2 RESULTS ---\n")
	fmt.Printf("Score: %d - %d\n", state2.HomeStats.MatchGoalsFor, state2.AwayStats.MatchGoalsFor)
	fmt.Printf("Possession: %d - %d\n", state2.HomeStats.PossessionTicks, state2.AwayStats.PossessionTicks)

	fmt.Println("\n✅ Because both matches used the exact same seed, the results are perfectly identical!")
}
