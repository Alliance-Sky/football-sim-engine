# RPS Football Engine

A high-performance, stateless Go simulation engine engineered specifically for **Multiplayer Football Manager Games** (e.g., Top Eleven, Hattrick).

This engine is designed to be integrated directly into live multiplayer backend architectures. It executes 90-minute matches in under `0.05` seconds, returning a comprehensive "delta payload" (`MatchState`) that can be instantly broadcasted over WebSockets for live frontend animations or persisted directly to your database.

---

## ⚡ What the Engine Does Automatically

As a backend developer, your primary responsibility is persisting the data before and after matches. The engine abstracts away all of the complex football simulation math:

### 1. Dynamic Performance Nerfs (Health & Fatigue)
You pass the player's current health from your database into the match.
* The engine calculates the health deficit (`100 - CurrentHealth`) and subtracts it directly from the player's Base Rating on every simulation tick.
* A 90-rated player with 75 Health automatically plays like a 65-rated player for the match duration.

### 2. Out-of-Position (OOP) Penalties
You can put any player in any position, but the engine applies granular rating penalties based on their `NaturalPosition` vs `AssignedPosition`:
* Checks footedness (e.g., a left-footed Left Back swapping to Right Back is penalized, but a both-footed player is not).
* Sector swaps (Defense to Midfield) are penalized heavily (~35%).
* Putting outfield players in Goal incurs a massive 50% rating penalty.

### 3. Tactical RPS (Rock-Paper-Scissors) & Home Advantage
Formations are grouped into tactical categories (Rock, Paper, Scissors):
* **Rock (Aggressive)**: 4-2-4, 3-4-3, 4-3-3 Attack
* **Paper (Balanced / Control)**: 4-2-3-1, 4-3-3 Hold, 3-5-2
* **Scissors (Counter / Defense)**: 5-3-2, 4-4-2 Flat, 5-4-1
* **Tactical Edge**: Winning the tactical matchup provides a numerical `Edge` (+1) buffing ball progression into the box and Expected Goals ($xG$).
* **Home Advantage**: Increases 50/50 loose ball retention and turnover recovery from **50% to 55%**.

### 4. Extra Time & Sudden Death Penalty Shootouts
* **League Matches (`NewLeagueMatch`)**: 90-minute regulation with draws allowed.
* **Cup Matches (`NewCupMatch`)**: If tied after 90 minutes, triggers Extra Time (91'-120'). If still tied, executes Best-of-5 penalties followed by round-by-round **Sudden Death** until a winner is determined.

### 5. Mandatory Modular Kit Colors
Teams require a mandatory non-empty `KitColor` (e.g., hex `#ff6b6b`) with zero engine hardcoding:
```go
team, err := NewTeam("t1", "Arsenal", "4-3-3 Attacking", "#f06595", players)
```

### 6. Canonical Post-Match Statistics Suite
Immediately after `The Final Whistle!`, the match log records a locked sequence of core statistics:
1. **Expected Goals (xG)**: Cumulative shot $xG$ based on defensive pressure.
2. **Possession**: Calculated directly from ticks in possession.
3. **Shots**: Total attempts on goal.
4. **Shots On Target (SOT)**: Accurate goal-bound strikes.
5. **Goalkeeper Saves**: Shots stopped by goalkeepers.
6. **Duels Won**: Defensive tackles and loose ball recoveries.
7. **Team Rating**: Historical overall team capability index.

### 7. Delta Payloads & JSON API Ready
The engine's outputs (`MatchState`, `ClubMatchStats`, `PlayerMatchStats`) are designed as "deltas" easily applied to a database. The engine natively tracks `Matches`, `Wins`, `Draws`, `Losses`, and `HealthLost` per match. Furthermore, all core structs are fully tagged for JSON (`json:"camelCase"`), and the `Commentary` timeline returns an array of `MatchLog` objects (pairing the `Minute` with the `Message`), making it trivial to pipe the simulation output directly into a frontend UI.

### 8. Multiplayer "Live Broadcast" Ready
Because the engine completes a full match simulation in milliseconds, it is perfectly suited for live multiplayer architectures. Your backend server can simulate a full matchday (e.g., at 00:00 UTC), instantly save the outcome to your database, and then broadcast the resulting JSON `Commentary` array over WebSockets. Tens of thousands of connected mobile/web clients can then receive the payload and animate the match timeline tick-by-tick to create a thrilling "live viewing" experience without stressing your backend servers.

---

## 🎨 Dedicated Vector SVG Assets (`svg-icons/`)

All match events, scoreboard kits, and post-match stats use dedicated, football-tailored SVG files located in `/svg-icons/`:

* `ball.svg` — Panel-patterned football
* `counterattack.svg` — Rapid midfield penetration break
* `save.svg` & `saves.svg` — Goalkeeper padded glove
* `tackle.svg` — Sliding tackle block
* `woodwork.svg` — Goal frame & crossbar strike
* `clearance.svg` — Defensive clearance shield
* `formation.svg` — Tactical pitch whiteboard layout
* `kickoff.svg` — Referee match kickoff whistle
* `final-whistle.svg` — Full-time triple whistle
* `xg.svg` — Goal frame with analytical $xG$ curve
* `shots.svg` — Football boot goal strike
* `shots-on-target.svg` — Ball inside crosshair target
* `possession.svg` — Ball control pitch quadrant
* `duels.svg` — 50/50 ground duel
* `team-rating.svg` — Football club crest badge with star
* `rock.svg`, `paper.svg`, `scissors.svg` — Tactical formation icons
* `home-advantage.svg` — Home stadium fortress
* `kit.svg` — Match jersey outline

---

## 💻 Core Developer Tools & Automators

The engine ships with a massive suite of developer tools designed to make building a backend trivial.

### 1. JSON Hydration (Database Loaders)
Never write a manual `for` loop to build a team again. If you pull raw JSON from your PostgreSQL/MongoDB database, the engine natively deserializes, hydrates, and validates the `*Team` struct in one line:
```go
jsonData := []byte(`[{"id": "p1", "name": "B. Saka", "naturalPosition": "RW", "rating": 88, ...}]`)
arsenal, err := engine.LoadTeamFromJSON("t1", "Arsenal", "4-3-3 Attacking", "#f06595", jsonData)
```

### 2. The LeagueManager (Season Automator)
The `LeagueManager` handles scheduling, math, points, and standings so you don't have to. It automatically calculates **Points**, **Goal Difference**, **Recent Form (W/D/L)**, and dynamically interleaves **Knockout Cup Matches** into your season!
```go
// Natively validate Rating Caps! (0, 0 = no caps). True = Double Round Robin.
league, _ := engine.NewLeagueManager(teams, 85.0, 0, true) 

// The engine natively tracks custom rewards for the winners!
league.PositionRewards[1] = map[string]int{"Prestige": 1000, "Gold": 5000}
league.PositionRewards[2] = map[string]int{"Prestige": 500, "Gold": 2000}

for round := 1; round <= len(league.Schedule); round++ {
    matchday := league.GetNextRound()
    for _, fixture := range matchday.Fixtures {
        state, _ := engine.QuickPlay(matchday.Type, fixture.Home, fixture.Away, true)
        league.RecordMatch(state)
    }
}

// 3. Output the perfectly sorted JSON Leaderboards directly to your frontend!
pointsTableJSON := league.GetTable()
goldenBootJSON := league.GetTopScorers(10)
```

### 3. TournamentManager (Sit-and-Go Tournaments)
Perfect for replicating classic knockout mechanics with 60-second "ready up" lobbies. It natively supports **Rating Caps**, dynamically pairs surviving teams, and automatically handles prize resolution.
```go
// Create a 16-team tournament with a U-80 Player Rating Cap
tourney, _ := engine.NewTournamentManager("tour-01", "Silver Cup", 16, 80.0, 0)
tourney.WinnerRewards["Gold_Chest"] = 1

tourney.Join(realPlayerTeam) // Dynamically nerfs any 80+ players down to exactly 80.0 for the simulation!

// If the lobby timer expires at 11/16 players, the engine fills the rest with perfectly scaled Bots!
tourney.StartWithBots() 

// The engine tracks eliminated teams and dynamically generates the bracket
for {
    fixtures, _ := tourney.GenerateNextRound()
    if len(fixtures) == 0 { break }
    
    // Simulate the round...
}

// Hook into the final results
winnerID := tourney.GetWinnerID()
rewards := tourney.WinnerRewards
```

### 4. Universal Bot Generation
If your League or Tournament lobby timer expires before it fills up, the engine can instantly pad your lobby with fully valid AI Teams. These bots pull from a massive dictionary of 200 cities, 200 first names, and 200 last names to feel completely realistic!
```go
// Automatically generates completely unique teams (e.g. "Paris Bot FC") 
// with players strictly tuned to your target rating (e.g. 75.0)
finalTeams := engine.FillWithBots(joinedTeams, 14, 75.0) 
league, _ := engine.NewLeagueManager(finalTeams, 0, 0)
```

### 5. Match Facades (QuickPlay & Deterministic)
* **`engine.QuickPlay(matchType, home, away, homeAdv)`**: A clean facade that auto-injects an RNG seed and returns the `MatchState` instantly.
* **`engine.DeterministicPlay(matchType, home, away, homeAdv, seed)`**: Replay classic matches or debug edge cases! If you pass the exact same seed (e.g., `9999`), the engine will perfectly recreate the exact same commentary, events, and scoreline every single time.

### 6. UI Dropdown Helpers
Building team management screens? The engine exports static arrays for your frontend:
```go
formations := engine.GetAvailableFormations() // ["4-3-3 Attacking", "4-2-3-1", ...]
positions := engine.GetPositions()            // ["GK", "CB", "CM", "ST", ...]
```

---

## 🚀 Usage

Check out the `/examples` directory for 4 standalone, heavily commented scripts demonstrating how to use `QuickPlay`, `DeterministicPlay`, `JSON Hydration`, and the `League Automator`!

To run the simulation tests and verify the underlying engine's logic:
```bash
go test ./engine/... -v
```
