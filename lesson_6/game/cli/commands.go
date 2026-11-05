package cli

import (
	"errors"
	"fmt"
	"strings"

	"golang_lessons/lesson_6/game/inventory"
)

// list печатает все предметы с их ID.
func (a *App) list(args []string) int {
	flags := a.newFlags("list")
	if code, stop := a.parse(flags, args); stop {
		return code
	}

	inv, ok := a.load()
	if !ok {
		return exitFailure
	}

	entries := inv.Entries()
	if len(entries) == 0 {
		fmt.Fprintln(a.stdout, "inventory is empty")
		return exitOK
	}
	for _, entry := range entries {
		fmt.Fprintf(a.stdout, "id=%d  %s\n", entry.ID, inventory.DescribeItem(entry.Item))
	}
	return exitOK
}

// addOptions — значения флагов команды add.
type addOptions struct {
	name       string
	weight     float64
	kind       string
	damage     int
	durability int
	defense    int
	effect     string
	charges    int
}

// add добавляет предмет и печатает его ID.
func (a *App) add(args []string) int {
	flags := a.newFlags("add")
	var opts addOptions
	flags.StringVar(&opts.name, "name", "", "item name")
	flags.Float64Var(&opts.weight, "weight", 0, "item weight")
	flags.StringVar(&opts.kind, "type", "weapon", "item type: weapon|armor|potion")
	flags.IntVar(&opts.damage, "damage", 5, "weapon damage")
	flags.IntVar(&opts.durability, "durability", 10, "weapon durability")
	flags.IntVar(&opts.defense, "defense", 5, "armor defense")
	flags.StringVar(&opts.effect, "effect", "+10 HP", "potion effect")
	flags.IntVar(&opts.charges, "charges", 3, "potion charges")
	if code, stop := a.parse(flags, args); stop {
		return code
	}

	if opts.name == "" {
		fmt.Fprintln(a.stderr, "flag -name is required")
		return exitUsage
	}
	item, err := buildItem(opts)
	if err != nil {
		// Все значения пришли из флагов, поэтому отказ (неверный тип, пустое или недопустимое
		// название, плохой вес) — это неверные флаги.
		fmt.Fprintln(a.stderr, err)
		return exitUsage
	}

	inv, ok := a.load()
	if !ok {
		return exitFailure
	}
	id, err := inv.AddItem(item)
	if err != nil {
		fmt.Fprintf(a.stderr, "add: %v\n", err)
		return exitFailure
	}
	if !a.save(inv) {
		return exitFailure
	}

	fmt.Fprintf(a.stdout, "added item id=%d\n", id)
	return exitOK
}

// use использует предмет по ID и сохраняет изменившееся состояние.
func (a *App) use(args []string) int {
	id, code, ok := a.parseID("use", args)
	if !ok {
		return code
	}

	inv, ok := a.load()
	if !ok {
		return exitFailure
	}
	item, err := inv.Get(id)
	if err != nil {
		return a.failLookup(id, err)
	}

	message, err := item.Use()
	switch {
	case errors.Is(err, inventory.ErrBroken):
		fmt.Fprintln(a.stderr, "cannot use: item is broken")
		return exitFailure
	case errors.Is(err, inventory.ErrExhausted):
		fmt.Fprintln(a.stderr, "cannot use: item is exhausted")
		return exitFailure
	case err != nil:
		fmt.Fprintf(a.stderr, "use error: %v\n", err)
		return exitFailure
	}

	// Успешное использование изменило предмет (прочность, заряды, надетость) — сохраняем.
	if !a.save(inv) {
		return exitFailure
	}
	fmt.Fprintln(a.stdout, message)
	return exitOK
}

// remove удаляет предмет по ID.
func (a *App) remove(args []string) int {
	id, code, ok := a.parseID("delete", args)
	if !ok {
		return code
	}

	inv, ok := a.load()
	if !ok {
		return exitFailure
	}
	if err := inv.Delete(id); err != nil {
		return a.failLookup(id, err)
	}
	if !a.save(inv) {
		return exitFailure
	}

	fmt.Fprintf(a.stdout, "deleted item id=%d\n", id)
	return exitOK
}

// buildItem создаёт предмет нужного типа из значений флагов add.
func buildItem(opts addOptions) (inventory.Item, error) {
	switch strings.ToLower(opts.kind) {
	case "weapon":
		weapon, err := inventory.NewWeapon(opts.name, opts.weight, opts.damage, opts.durability)
		if err != nil {
			return nil, err
		}
		return weapon, nil
	case "armor":
		armor, err := inventory.NewArmor(opts.name, opts.weight, opts.defense)
		if err != nil {
			return nil, err
		}
		return armor, nil
	case "potion":
		potion, err := inventory.NewPotion(opts.name, opts.weight, opts.effect, opts.charges)
		if err != nil {
			return nil, err
		}
		return potion, nil
	}
	return nil, fmt.Errorf("unknown type %q (use weapon|armor|potion)", opts.kind)
}
