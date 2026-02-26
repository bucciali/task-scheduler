package main

import (
	"final_project/pkg/api"
	"final_project/pkg/db"
	"fmt"
	"net/http"
	"os"
)

func main() {
	dbFile := "./dbSql/scheduler.db"
	webDir := "./web"
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}
	db.Init(dbFile)
	defer db.Db.Close()
	api.Init()
	fl := http.FileServer(http.Dir(webDir))
	http.Handle("/", fl)
	adress := fmt.Sprintf(":%s", port)
	fmt.Print("Server is running on port localhost", adress)
	err := http.ListenAndServe(adress, nil)
	if err != nil {
		fmt.Printf("Problems with starting server: %v", err)
	}

}
