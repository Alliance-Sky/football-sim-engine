package main

import (
	"encoding/json"
	"fmt"
	"log"

	"rps-football-engine/engine"
)

// Helper to quickly generate a real player team
func createRealTeam(id, name string, rating float64) *engine.Team {
	return engine.GenerateBotTeam(id, name, rating) // Using bot generator just for quick dummy data
}

func main() {
	fmt.Println("🚀 Starting Sit-and-Go Tournament (U-80 Rating Cap)")

	// 1. Create a 16-player tournament with an 80.0 Player Rating Cap
	tourney, err := engine.NewTournamentManager("tour-001", "Silver Cup", 16, 80.0, 0)
	if err != nil {
		log.Fatalf("Failed to create tournament: %v", err)
	}

	// 2. Real Players join the lobby
	fmt.Println("⏳ Waiting for players to join...")
	
	realTeam1 := createRealTeam("t1", "Real Player FC", 79.0)
	_ = tourney.Join(realTeam1)
	
	realTeam2 := createRealTeam("t2", "Too Good FC", 85.0) // This team has 85 rated players
	_ = tourney.Join(realTeam2)
	fmt.Printf("⚠️ Too Good FC joined with 85 OVR. The engine dynamically nerfed them to 80.0!\n")

	realTeam3 := createRealTeam("t3", "Underdog FC", 70.0)
	_ = tourney.Join(realTeam3)

	fmt.Printf("👤 Players joined: %d/%d\n", len(tourney.Participants), tourney.MaxParticipants)

	// 3. Timer expires, start with bots!
	fmt.Println("🤖 Timer expired! Filling remaining slots with AI Bots...")
	tourney.StartWithBots()
	fmt.Printf("✅ Tournament Started with %d teams!\n\n", len(tourney.ActiveTeams))

	// 4. The Knockout Loop
	for {
		fixtures, err := tourney.GenerateNextRound()
		if err != nil || len(fixtures) == 0 {
			break
		}

		fmt.Printf("--- ROUND %d (%d Fixtures) ---\n", tourney.RoundNumber, len(fixtures))
		
		// In a real backend, you would Wait 60 Seconds here before simulating!
		// time.Sleep(60 * time.Second) 

		for _, f := range fixtures {
			// Simulate the MatchCup
			state, _ := engine.QuickPlay(engine.MatchCup, f.Home, f.Away, false)
			tourney.RecordMatch(state)
			fmt.Printf("🏆 %s %d - %d %s\n", f.Home.Name, state.HomeStats.GoalsFor, state.AwayStats.GoalsFor, f.Away.Name)
		}
		fmt.Println()
	}

	// 5. Crowning the Champion
	fmt.Println("🎉 TOURNAMENT FINISHED!")
	fmt.Printf("🥇 WINNER: %s\n", tourney.Winner.Name)
	fmt.Printf("🥈 RUNNER UP: %s\n", tourney.RunnerUp.Name)

	fmt.Println("\n👟 GOLDEN BOOT WINNER:")
	scorers := tourney.GetTopScorers(1)
	if len(scorers) > 0 {
		scorerJSON, _ := json.MarshalIndent(scorers[0], "", "  ")
		fmt.Println(string(scorerJSON))
	}
}
