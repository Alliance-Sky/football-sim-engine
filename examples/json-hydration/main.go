package main

import (
	"encoding/json"
	"fmt"
	"log"

	"rps-football-engine/engine"
)

// Raw JSON string simulating a database fetch for a team's roster
const arsenalJSON = `[
	{"id": "p1", "name": "D. Raya", "naturalPosition": "GK", "foot": "Right", "rating": 85, "age": 28, "health": 100, "assignedPosition": "GK"},
	{"id": "p2", "name": "W. Saliba", "naturalPosition": "CB", "foot": "Right", "rating": 88, "age": 23, "health": 95, "assignedPosition": "CB"},
	{"id": "p3", "name": "Gabriel M.", "naturalPosition": "CB", "foot": "Left", "rating": 86, "age": 26, "health": 92, "assignedPosition": "CB"},
	{"id": "p4", "name": "O. Zinchenko", "naturalPosition": "LB", "foot": "Left", "rating": 84, "age": 27, "health": 100, "assignedPosition": "LB"},
	{"id": "p5", "name": "B. White", "naturalPosition": "RB", "foot": "Right", "rating": 85, "age": 26, "health": 100, "assignedPosition": "RB"},
	{"id": "p6", "name": "D. Rice", "naturalPosition": "DM", "foot": "Right", "rating": 89, "age": 25, "health": 100, "assignedPosition": "DM"},
	{"id": "p7", "name": "M. Ødegaard", "naturalPosition": "CM", "foot": "Left", "rating": 89, "age": 25, "health": 90, "assignedPosition": "CM"},
	{"id": "p8", "name": "K. Havertz", "naturalPosition": "CM", "foot": "Left", "rating": 85, "age": 24, "health": 100, "assignedPosition": "CM"},
	{"id": "p9", "name": "G. Martinelli", "naturalPosition": "LW", "foot": "Right", "rating": 86, "age": 22, "health": 100, "assignedPosition": "LW"},
	{"id": "p10", "name": "B. Saka", "naturalPosition": "RW", "foot": "Left", "rating": 88, "age": 22, "health": 98, "assignedPosition": "RW"},
	{"id": "p11", "name": "G. Jesus", "naturalPosition": "ST", "foot": "Right", "rating": 85, "age": 27, "health": 100, "assignedPosition": "ST"}
]`

const cityJSON = `[
	{"id": "c1", "name": "Ederson", "naturalPosition": "GK", "foot": "Left", "rating": 88, "age": 30, "health": 100, "assignedPosition": "GK"},
	{"id": "c2", "name": "R. Dias", "naturalPosition": "CB", "foot": "Right", "rating": 89, "age": 27, "health": 100, "assignedPosition": "CB"},
	{"id": "c3", "name": "J. Stones", "naturalPosition": "CB", "foot": "Right", "rating": 86, "age": 29, "health": 100, "assignedPosition": "CB"},
	{"id": "c4", "name": "J. Gvardiol", "naturalPosition": "LB", "foot": "Left", "rating": 86, "age": 22, "health": 100, "assignedPosition": "LB"},
	{"id": "c5", "name": "K. Walker", "naturalPosition": "RB", "foot": "Right", "rating": 85, "age": 33, "health": 90, "assignedPosition": "RB"},
	{"id": "c6", "name": "Rodri", "naturalPosition": "DM", "foot": "Right", "rating": 91, "age": 27, "health": 100, "assignedPosition": "DM"},
	{"id": "c7", "name": "K. De Bruyne", "naturalPosition": "CM", "foot": "Right", "rating": 91, "age": 32, "health": 85, "assignedPosition": "CM"},
	{"id": "c8", "name": "Bernardo S.", "naturalPosition": "CM", "foot": "Left", "rating": 88, "age": 29, "health": 100, "assignedPosition": "CM"},
	{"id": "c9", "name": "J. Grealish", "naturalPosition": "LW", "foot": "Right", "rating": 85, "age": 28, "health": 100, "assignedPosition": "LW"},
	{"id": "c10", "name": "P. Foden", "naturalPosition": "RW", "foot": "Left", "rating": 88, "age": 23, "health": 100, "assignedPosition": "RW"},
	{"id": "c11", "name": "E. Haaland", "naturalPosition": "ST", "foot": "Left", "rating": 91, "age": 23, "health": 100, "assignedPosition": "ST"}
]`

func main() {
	fmt.Println("🚀 Hydrating teams directly from JSON...")

	// 1. One-line hydration of the entire team from raw JSON bytes
	arsenal, err := engine.LoadTeamFromJSON("t1", "Arsenal", "4-3-3 Attacking", "#f06595", []byte(arsenalJSON))
	if err != nil {
		log.Fatalf("Failed to load Arsenal: %v", err)
	}

	manCity, err := engine.LoadTeamFromJSON("t2", "Man City", "4-3-3 Attacking", "#51cf66", []byte(cityJSON))
	if err != nil {
		log.Fatalf("Failed to load Man City: %v", err)
	}

	fmt.Printf("✅ Arsenal loaded successfully! Overall Rating: %.1f\n", arsenal.OverallRating())
	fmt.Printf("✅ Man City loaded successfully! Overall Rating: %.1f\n", manCity.OverallRating())

	// 2. Play Match
	fmt.Println("\n⚽ Simulating match...")
	state, _ := engine.QuickPlay(engine.MatchLeague, arsenal, manCity, true)

	fmt.Printf("\n--- FULL TIME ---\n")
	fmt.Printf("%s %d : %d %s\n\n", arsenal.Name, state.HomeStats.MatchGoalsFor, state.AwayStats.MatchGoalsFor, manCity.Name)

	fmt.Println("📦 First 3 Match Log Events:")
	outputLogs, _ := json.MarshalIndent(state.Commentary[:3], "", "  ")
	fmt.Println(string(outputLogs))
}
