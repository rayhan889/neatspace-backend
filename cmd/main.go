package main

import (
	"fmt"
	"os"

	"github.com/rayhan889/neatspace/cmd/commands"
)

// @title	    Neatspace API
// @description	Neatspace API documentation
// @version		1.0

// @securityDefinitions.http bearerAuth
// @scheme bearer
// @bearerFormat JWT

// @host      localhost:8080
// @BasePath  /api/v1
func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
