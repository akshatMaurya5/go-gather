package main

import (
	"go-gather/db"
)

func main() {
	// log.Print("Starting WebSocket server...")

	// wsh := ws.WebSocketHandler{
	// 	Upgrader: websocket.Upgrader{
	// 		CheckOrigin: func(r *http.Request) bool { return true },
	// 	},
	// }

	// http.HandleFunc("/ws", wsh.ServeHTTP)

	// log.Print("WebSocket server listening on port 3001...")

	db.TestSQLDbConnection()
	// log.Fatal(http.ListenAndServe(":3001", nil))
}
