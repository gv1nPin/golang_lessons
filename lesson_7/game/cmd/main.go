package main

import (
	"os"

	"golang_lessons/lesson_7/game/cli"
	"golang_lessons/lesson_7/game/storage"
)

// dataFile — файл с инвентарём. Путь выбирается здесь, при сборке зависимостей.
const dataFile = "inventory.dat"

// main только собирает зависимости и завершает процесс с кодом команды.
// os.Exit вызывается здесь и больше нигде.
func main() {
	store := storage.NewFileStore(dataFile)
	app := cli.New(store, os.Stdout, os.Stderr)
	os.Exit(app.Run(os.Args[1:]))
}
