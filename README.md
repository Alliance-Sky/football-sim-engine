# RPS Football Engine

![Go Version](https://img.shields.io/github/go-mod/go-version/rps-football-engine/rps-football-engine)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

A high-performance, stateless Go simulation engine engineered specifically for **Multiplayer Football Manager Games** (e.g., Top Eleven, Hattrick, Football Manager).

Designed to be integrated directly into live multiplayer backend architectures, this engine executes full 90-minute matches (including extra time and penalties) in under `0.05` seconds. It returns a comprehensive delta payload (`MatchState`) that can be broadcasted via WebSockets for live frontend animations, or persisted directly to your database.

---

## 🚀 Key Features

### 1. ⚡ High-Performance Core Engine
* **Stateless & Tick-Based:** Simulates a match tick-by-tick instantly without side effects, making it entirely thread-safe and perfect for high-throughput backends.
* **Dual Output JSON Payloads:** Post-match stats for both **Players** and **Clubs** brilliantly return both **Single-Match Deltas** (e.g., `matchGoalsFor: 3`) for animating UI popups, AND **Cumulative Lifetime Stats** (e.g., `postMatchWins: 14`, `postMatchHealth: 93.5`) so you can blindly execute `UPDATE` queries against your database without doing any math.
* **Deterministic Mode:** By supplying an RNG seed, you can flawlessly reproduce the exact same match events and outcome every single time—ideal for debugging and replay systems.

### 2. 🧠 Advanced Football Logic
* **Dynamic Player Progression:** Every player has a `Potential` ceiling. Post-match, players below their potential have a 10% RNG chance to permanently gain +1.0 Rating in-memory, simulating "wonderkid" growth.

* **Automated In-Memory Fatigue:** The engine automatically deducts calculated fatigue straight from the `Player.Health` struct in-memory. If you feed the exact same `Team` struct back into the engine for Match 2, their fatigue perfectly carries over naturally.
* **Out-of-Position (OOP) Penalties:** Automatically applies granular rating penalties based on a player's `NaturalPosition` vs `AssignedPosition` (e.g., sector swaps, wrong foot, outfield players in goal).
* **Tactical Rock-Paper-Scissors (RPS):** Formations fall into categorical advantages (e.g., Aggressive > Control > Counter > Aggressive), conferring statistical edges in-game.
* **In-Depth Match Resolvers:** Features complex mathematical models for Progression, Creation, and Defensive Shields, dynamically affecting Ball Zones and generating realistic Expected Goals ($xG$).

### 3. 🏆 Tournaments & Season Automators
* **LeagueManager:** Fully automates season scheduling with round-robin fixtures interleaved with a Knockout Cup. Handles lean, perfectly sorted standings (Points > GD > GF), seamless cumulative player season stats, and end-of-season rewards.
* **TournamentManager (Sit-and-Go):** Manages quick-fire knockout brackets, tracks eliminated teams, automatically handles Two-Legged aggregates, and pairs opponents dynamically.
* **GroupTournamentManager (World Cup Format):** Orchestrates large-scale Group Stages that mathematically split 32 teams into groups, transitioning the top teams into a 16-team knockout bracket.
* **Universal Bots:** Generate fully functioning, statistically balanced AI teams dynamically to fill empty tournament slots.

---

## 📦 Project Structure

```text
football-sim-engine/
├── engine/              # Core football simulation engine packages
│   ├── bots.go          # AI Bot and dummy team generation
│   ├── config.go        # Constants, Formations, and XG configurations
│   ├── engine.go        # Team & Player data structures, stat aggregations
│   ├── league.go        # LeagueManager and Season round-robin automators
│   ├── matches.go       # Match lifecycle, extra time, penalty shootouts
│   ├── models.go        # Player models, MatchState, and JSON structures
│   ├── resolvers.go     # Tick-by-tick event logic (Passes, Tackles, Shots)
│   ├── stats.go         # Play-by-play stats tracking (Goals, Assists)
│   └── tournament.go    # Knockout & Sit-and-Go bracket management
└── examples/            # Example scripts demonstrating core integrations
    ├── basic-match/             # Standard 90-minute match QuickPlay
    ├── deterministic-match/     # Seeded repeatable matches
    ├── json-hydration/          # Loading teams from DB JSON
    ├── league-automator/        # Running a full multi-team season
    ├── sit-and-go/              # 16-team knockout bracket with bots
    └── websocket-broadcaster/   # Broadcasting MatchState commentary over WS
```

---

## 💻 Getting Started

### Installation

```bash
go get github.com/yourusername/rps-football-engine/engine
```

### Quick Match Simulation

Using the engine is extremely straightforward. Here is an example of running a basic match:

```go
package main

import (
	"encoding/json"
	"fmt"
	"rps-football-engine/engine"
)

func main() {
    // 1. Load Teams (or hydrate directly from JSON)
    homeTeam, _ := engine.NewTeam("t1", "Arsenal", "4-3-3 Attacking", "#f06595", getPlayers())
    awayTeam, _ := engine.NewTeam("t2", "Chelsea", "4-2-3-1", "#034694", getPlayers())
    
    // 2. Play Match (League Match = 90 mins, No Extra Time)
    matchState, err := engine.QuickPlay(engine.MatchLeague, homeTeam, awayTeam, true) // true = Home Advantage
    if err != nil {
        panic(err)
    }
    
    // 3. Print Results or serialize to frontend
    fmt.Printf("Final Score: %d - %d\n", matchState.HomeStats.MatchGoalsFor, matchState.AwayStats.MatchGoalsFor)
    
    // Convert the entire state delta to JSON for your frontend
    jsonBytes, _ := json.MarshalIndent(matchState, "", "  ")
    fmt.Println(string(jsonBytes))
}
```

---

## 🛠️ Deep Dive: Core Orchestrators

### The League Automator

No more manual looping to build seasons. `LeagueManager` tracks points, goal differences, and seamlessly handles all cumulative player stats and interleaved knockout cups!

```go
// 2 = Double Round Robin (Home and Away)
league, _ := engine.NewLeagueManager(teams, 85.0, 0, 2) 

for round := 1; round <= len(league.Schedule); round++ {
    matchday := league.GetNextRound()
    for _, fixture := range matchday.Fixtures {
        // Because the engine uses in-memory pointers (*Player), fatigue and 
        // lifetime stats naturally carry over from match to match seamlessly!
        state, _ := engine.QuickPlay(matchday.Type, fixture.Home, fixture.Away, true)
        league.RecordMatch(state)
    }
}

// Extract JSON leaderboard for your clients
standingsJSON := league.GetTable()
topScorersJSON := league.GetTopScorers(10)
```

### JSON Hydration (Database Loaders)

Never write manual mappers. If you pull raw JSON from PostgreSQL/MongoDB, the engine natively deserializes, hydrates, and validates the `*Team` struct:

```go
jsonData := []byte(`[{"id": "p1", "name": "B. Saka", "naturalPosition": "RW", "rating": 88, ...}]`)
arsenal, err := engine.LoadTeamFromJSON("t1", "Arsenal", "4-3-3 Attacking", "#f06595", jsonData)
```

### Multiplayer WebSockets "Live Broadcast"

Because the engine evaluates an entire match essentially instantaneously, it is perfect for live multiplayer. 

**Standard Flow:**
1. A centralized Execution Worker queries your DB queue for scheduled matches.
2. The worker runs `engine.QuickPlay(...)` taking `<0.05` seconds.
3. The worker saves the results (`postMatchWins`, `postMatchGoalsFor`, `postMatchHealth`) to the database immediately.
4. The worker broadcasts the `MatchState.Commentary` array over WebSockets.
5. Thousands of connected mobile/web clients animate the timeline tick-by-tick, creating a thrilling "live viewing" experience without placing any rendering load on the backend.

---

## 🧪 Testing

To run the simulation tests and verify the underlying engine's logic, use the standard Go toolchain:

```bash
go test ./engine/... -v
```

*Note: For further integration examples, dive into the `/examples` directory.*
