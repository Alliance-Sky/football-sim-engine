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
				25, 100.0, new(engine.Position), 90.0,
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

	// Let's pretend only 2 real players joined a 4-team league. 
	// We will fill the rest with Bots!
	arsenal := createDummyTeam("t1", "Arsenal", "4-3-3 Attacking", "#ff0000", 8)
	manCity := createDummyTeam("t2", "Man City", "4-2-3-1", "#0000ff", 7)
	teams := []*engine.Team{arsenal, manCity}
	teams = engine.FillWithBots(teams, 4, 75.0)

	// 2. Initialize the League Automator with no caps, and Double Round-Robin enabled
	league, err := engine.NewLeagueManager(teams, 0, 0, 2)
	if err != nil {
		log.Fatalf("Failed to create league: %v", err)
	}

	fmt.Printf("📅 Schedule generated: %d rounds total.\n\n", len(league.Schedule))

	// 3. Simulate the entire season
	for round := 1; round <= len(league.Schedule); round++ {
		r := league.GetNextRound()
		
		fmt.Printf("Playing Matchday %d (%s) - %d fixtures\n", round, r.Type, len(r.Fixtures))
		
		for _, f := range r.Fixtures {
			// Simulate the match (Automatically applies Cup logic like Extra Time if needed)
			state, _ := engine.QuickPlay(r.Type, f.Home, f.Away, true)
			
			// Feed the result back into the league manager!
			league.RecordMatch(state)
			
			if r.Type == engine.MatchCup {
				fmt.Printf("  🏆 CUP RESULT: %s %d - %d %s (Winner: %v)\n", f.Home.Name, state.HomeStats.MatchGoalsFor, state.AwayStats.MatchGoalsFor, f.Away.Name, *state.Winner)
			}
		}
	}

	// 4. Output the generated standings
	fmt.Println("\n📊 FINAL LEAGUE STANDINGS (Before Deductions):")
	table := league.GetTable()
	for i, standing := range table {
		fmt.Printf("%d. %s - %d pts (GD: %d)\n", i+1, standing.TeamName, standing.Points, standing.GoalDifference)
	}

	// --- NEW FEATURE SHOWCASE ---
	// Deduct 10 points from the Champion for Financial Fair Play violations!
	championID := league.GetChampionID()
	fmt.Printf("\n🚨 BREAKING NEWS: %s has been deducted 10 points for FFP violations!\n", table[0].TeamName)
	_ = league.DeductPoints(championID, 10)

	fmt.Println("\n📊 UPDATED LEAGUE STANDINGS:")
	updatedTable := league.GetTable() // Auto-resorts!
	for i, standing := range updatedTable {
		fmt.Printf("%d. %s - %d pts (GD: %d)\n", i+1, standing.TeamName, standing.Points, standing.GoalDifference)
	}

	// Check if Man City has any future fixtures (they shouldn't, season is over)
	// But during the season, developers can use this to show users their Calendar!
	mancitySchedule := league.GetTeamSchedule("t2")
	fmt.Printf("\n📅 Man City Upcoming Fixtures: %d\n", len(mancitySchedule))
	// ----------------------------

	// 5. Output Golden Boot Winner
	fmt.Println("\n👟 GOLDEN BOOT WINNER:")
	scorers := league.GetTopScorers(1)
	if len(scorers) > 0 {
		scorerJSON, _ := json.MarshalIndent(scorers[0], "", "  ")
		fmt.Println(string(scorerJSON))
	}
}
