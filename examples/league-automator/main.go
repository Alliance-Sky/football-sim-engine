package main

import (
	"encoding/json"
	"fmt"
	"log"

	"rps-football-engine/engine"
)

func createDummyTeam(id, name, formation, kit string, offset int) *engine.Team {
	slots := engine.FormationSlots[formation]
	var players []*engine.Player
	playerIdx := 1

	for pos, count := range slots {
		for i := 0; i < count; i++ {
			p, _ := engine.NewPlayer(
				fmt.Sprintf("%s-p%d", id, playerIdx),
				fmt.Sprintf("%s Player %d", name, playerIdx),
				pos, engine.FootRight,
				float64(80+offset), // Rating varies by offset
				25, 100.0, new(engine.Position),
			)
			p.AssignedPosition = pos
			players = append(players, p)
			playerIdx++
		}
	}

	team, _ := engine.NewTeam(id, name, formation, kit, players)
	return team
}

func main() {
	fmt.Println("🚀 Initializing League Automator...")

	// 1. Create 4 Teams for a mini-league
	teams := []*engine.Team{
		createDummyTeam("t1", "Arsenal", "4-3-3 Attacking", "#ff0000", 8), // Strongest
		createDummyTeam("t2", "Man City", "4-2-3-1", "#0000ff", 7),
		createDummyTeam("t3", "Liverpool", "4-3-3 Attacking", "#00ff00", 6),
		createDummyTeam("t4", "Chelsea", "3-5-2", "#000000", 5), // Weakest
	}

	// 2. Initialize the League Manager
	league, err := engine.NewLeagueManager(teams)
	if err != nil {
		log.Fatalf("Failed to create league: %v", err)
	}

	fmt.Printf("📅 Schedule generated: %d rounds total.\n\n", len(league.Schedule))

	// 3. Simulate the entire season
	for round := 1; round <= len(league.Schedule); round++ {
		fixtures := league.GetNextRound()
		for _, f := range fixtures {
			// Simulate the match
			state, _ := engine.QuickPlay(engine.MatchLeague, f.Home, f.Away, true)
			
			// Feed the result back into the league manager!
			league.RecordMatch(state)
		}
	}

	// 4. Output the Final League Table
	fmt.Println("🏆 FINAL LEAGUE TABLE:")
	table := league.GetTable()
	tableJSON, _ := json.MarshalIndent(table, "", "  ")
	fmt.Println(string(tableJSON))

	// 5. Output Golden Boot Winner
	fmt.Println("\n👟 GOLDEN BOOT WINNER:")
	scorers := league.GetTopScorers(1)
	if len(scorers) > 0 {
		scorerJSON, _ := json.MarshalIndent(scorers[0], "", "  ")
		fmt.Println(string(scorerJSON))
	}
}
