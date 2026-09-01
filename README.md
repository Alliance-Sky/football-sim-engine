# RPS Football Engine

A pure, side-effect-free Go simulation engine for football (soccer) matches.

This engine is designed to be integrated into backend game architectures. It requires you to pass in unmutated player and team data, and returns a comprehensive "delta payload" (`MatchState`) that you can easily persist to your database.

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

## 💻 Basic Implementation Example

```go
import "rps-football-engine/engine"

// 1. Load Player Lineups
p1, _ := engine.NewPlayer("uuid-001", "Bukayo Saka", engine.PosRW, engine.FootLeft, 88.0, 23, 100.0, nil)
// ... instantiate full starting XI

// 2. Initialize Teams with Mandatory KitColor
homeTeam, _ := engine.NewTeam("t1", "Arsenal", "4-3-3 Attacking", "#f06595", homePlayers)
awayTeam, _ := engine.NewTeam("t2", "Man City", "4-2-3-1", "#51cf66", awayPlayers)

// 3. Play Match (League or Cup)
match, _ := engine.NewLeagueMatch(homeTeam, awayTeam, true, true)
state := match.Play()

// 4. Persist Results & Delta Payloads
// The engine provides complete deltas (Matches, Wins, Goals, etc.) to update your DB cleanly.
db.Execute("UPDATE clubs SET matches = matches + ?, wins = wins + ?, draws = draws + ?, losses = losses + ? WHERE id = ?",
    state.HomeStats.Matches, state.HomeStats.Wins, state.HomeStats.Draws, state.HomeStats.Losses, state.HomeStats.Team.ID)
    
for _, stats := range state.PlayerStats.Stats {
    if stats.Appearances > 0 {
        db.Execute("UPDATE players SET matches = matches + 1, health = health - ? WHERE id = ?", 
            stats.HealthLost, stats.Player.ID)
    }
}
```

---

## 🚀 Usage

The engine is designed as a standalone Go module to be imported into your backend services. It does not include a standalone web server.

To run the simulation tests and verify the engine's logic:
```bash
go test ./engine/... -v
```
