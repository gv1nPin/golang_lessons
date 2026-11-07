// Package cli разбирает аргументы командной строки и выполняет команды инвентаря.
// Пакет получает хранилище и потоки вывода снаружи и сам ничего не создаёт и не завершает процесс.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"golang_lessons/lesson_7/game/inventory"
	"golang_lessons/lesson_7/game/storage"
)

// Store — хранилище инвентаря, которое нужно командам. Интерфейс объявлен здесь, со стороны
// потребителя, и содержит только то, чем пользуются команды. Его реализуют и storage.FileStore,
// и storage.MemoryStore, и любая тестовая заглушка.
type Store interface {
	Load() (*inventory.Inventory, error)
	Save(inv *inventory.Inventory) error
}

// Коды завершения.
const (
	exitOK      = 0 // команда выполнена
	exitFailure = 1 // операционная ошибка: хранилище недоступно или повреждено, предмет не найден или не годится
	exitUsage   = 2 // неверное использование: неизвестная команда, неверные или пропущенные флаги
)

// App выполняет команды инвентаря над переданным хранилищем.
type App struct {
	store  Store
	stdout io.Writer
	stderr io.Writer
}

// New создаёт приложение. Результаты команд пишутся в stdout, ошибки в stderr.
func New(store Store, stdout, stderr io.Writer) *App {
	return &App{store: store, stdout: stdout, stderr: stderr}
}

// Run выбирает команду по args и выполняет её. Возвращает код завершения;
// сам процесс не завершает: os.Exit вызывается только в main.
func (a *App) Run(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(a.stderr, "usage: inventory <command> [flags]")
		return exitUsage
	}

	commands := map[string]func([]string) int{
		"list":   a.list,
		"add":    a.add,
		"use":    a.use,
		"delete": a.remove,
	}
	// Команда проверяется до обращения к хранилищу: неизвестная команда даёт код 2,
	// даже если хранилище недоступно или повреждено.
	command, ok := commands[args[0]]
	if !ok {
		fmt.Fprintf(a.stderr, "unknown command: %s\n", args[0])
		return exitUsage
	}
	return command(args[1:])
}

// newFlags создаёт набор флагов подкоманды. У каждой подкоманды он свой.
// Вывод самого пакета flag отключён: ошибки печатаем мы, один раз и в stderr.
func (a *App) newFlags(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	return flags
}

// parse разбирает флаги и отклоняет лишние позиционные аргументы.
// Если stop == true, команду выполнять не нужно: надо вернуть code.
func (a *App) parse(flags *flag.FlagSet, args []string) (code int, stop bool) {
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			flags.SetOutput(a.stdout)
			flags.PrintDefaults()
			return exitOK, true
		}
		fmt.Fprintln(a.stderr, err)
		return exitUsage, true
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(a.stderr, "unexpected positional arguments for %s: %v\n", flags.Name(), flags.Args())
		return exitUsage, true
	}
	return 0, false
}

// parseID разбирает команды, которым нужен только флаг -id (use и delete).
// Правило «ID обязателен и положителен» живёт здесь, в одном месте.
func (a *App) parseID(command string, args []string) (id int, code int, ok bool) {
	flags := a.newFlags(command)
	flags.IntVar(&id, "id", 0, "item id")
	if code, stop := a.parse(flags, args); stop {
		return 0, code, false
	}
	if id <= 0 {
		fmt.Fprintln(a.stderr, "flag -id is required and must be positive")
		return 0, exitUsage, false
	}
	return id, 0, true
}

// load загружает инвентарь. При ошибке печатает её в stderr и возвращает false;
// для повреждённых данных достаёт из ошибки номер и содержимое строки через errors.As.
func (a *App) load() (*inventory.Inventory, bool) {
	inv, err := a.store.Load()
	if err != nil {
		fmt.Fprintf(a.stderr, "load inventory: %v\n", err)
		var parseErr *storage.ParseItemError
		if errors.As(err, &parseErr) {
			fmt.Fprintf(a.stderr, "  line %d: %q\n", parseErr.Line, parseErr.Value)
		}
		return nil, false
	}
	return inv, true
}

// save сохраняет инвентарь. При ошибке печатает её в stderr и возвращает false.
func (a *App) save(inv *inventory.Inventory) bool {
	if err := a.store.Save(inv); err != nil {
		fmt.Fprintf(a.stderr, "save: %v\n", err)
		return false
	}
	return true
}

// failLookup печатает ошибку поиска предмета по ID и возвращает код 1.
// Общая для use и delete: правило «предмета с таким ID нет» одно.
func (a *App) failLookup(id int, err error) int {
	if errors.Is(err, inventory.ErrNotFound) {
		fmt.Fprintf(a.stderr, "item id=%d not found\n", id)
	} else {
		fmt.Fprintf(a.stderr, "item id=%d: %v\n", id, err)
	}
	return exitFailure
}
