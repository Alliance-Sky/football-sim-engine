package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"rps-football-engine/engine"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for demo
	},
}

// simulateLiveMatch runs the engine instantly, but streams the commentary slowly.
func simulateLiveMatch(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// 1. Create Teams
	home := createBotTeam("t1", "Real Madrid", 85.0)
	away := createBotTeam("t2", "Barcelona", 84.0)

	// 2. The Engine runs the entire match INSTANTLY (2 milliseconds)
	log.Println("Generating Match Data...")
	state, err := engine.QuickPlay(engine.MatchLeague, home, away, true)
	if err != nil {
		log.Println("Engine error:", err)
		return
	}
	
	// 3. We have the full result immediately! We can save to DB here.
	log.Printf("Match Generated. Final Score: %d - %d", state.HomeStats.GoalsFor, state.AwayStats.GoalsFor)

	// 4. We start broadcasting the timeline to the client to create a "Live" feel
	for _, event := range state.Commentary {
		// Send the MatchLog struct as JSON over WebSocket
		err := conn.WriteJSON(event)
		if err != nil {
			log.Println("Client disconnected:", err)
			return
		}
		
		// Wait 250 milliseconds between events so the UI animates properly
		time.Sleep(250 * time.Millisecond)
	}

	// 5. Send a final termination message
	finalMsg := engine.MatchLog{
		Minute:  90,
		Message: fmt.Sprintf("FULL TIME! The final score is %d - %d.", state.HomeStats.GoalsFor, state.AwayStats.GoalsFor),
		Type:    engine.LogNeutral,
	}
	conn.WriteJSON(finalMsg)
	log.Println("Broadcast complete.")
}

func main() {
	// Simple static HTML frontend to connect to the WebSocket
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// The WebSocket endpoint
	http.HandleFunc("/ws", simulateLiveMatch)

	fmt.Println("📺 Live Broadcast Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createBotTeam(id, name string, rating float64) *engine.Team {
	teams := []*engine.Team{
		{ID: id, Name: name, Formation: "4-3-3 Attacking"},
	}
	return engine.FillWithBots(teams, 1, rating)[0]
}
