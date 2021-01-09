package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Print("Migration name: ")
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		panic(err)
	}

	migrationName := strings.Join(strings.Split(input, " "), "_")
	date := time.Now().Format("20060102150405")
	content := []byte("-- +migrate Up\n\n-- +migrate Down\n\n")

	err = ioutil.WriteFile(fmt.Sprintf("db/migrations/%s_%s.sql", date, migrationName), content, 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Migration created")
}
