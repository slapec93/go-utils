package main

import (
	"fmt"
	"os"

	"github.com/slapec93/go-utils/pkg/database"
)

func main() {
	fmt.Println("Setup db connection for migrations ...")
	err := database.RunMigrations()
	if err != nil {
		fmt.Printf("Migration failed: %s", err)
		os.Exit(1)
	}

	fmt.Println("Migration finished ...")
}
