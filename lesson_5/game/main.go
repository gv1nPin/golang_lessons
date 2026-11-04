package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	ErrBroken    = errors.New("item is broken")
	ErrExhausted = errors.New("item is exhausted")
	ErrNotFound  = errors.New("item not found")
)

// ParseItemError — типизированная ошибка разбора строки инвентаря.
type ParseItemError struct {
	Line  int
	Value string
	Err   error
}

func (e *ParseItemError) Error() string {
	return fmt.Sprintf("parse error on line %d %q: %v", e.Line, e.Value, e.Err)
}

func (e *ParseItemError) Unwrap() error {
	return e.Err
}

type Item interface {
	Name() string
	Weight() float64
	Use() (string, error)
	ID() int
	SetID(id int)
}

var (
	_ Item = (*Weapon)(nil)
	_ Item = (*Armor)(nil)
	_ Item = (*Potion)(nil)
)

type Weapon struct {
	id         int
	name       string
	weight     float64
	Damage     int
	Durability int
}

func (w *Weapon) Name() string    { return w.name }
func (w *Weapon) Weight() float64 { return w.weight }
func (w *Weapon) ID() int         { return w.id }
func (w *Weapon) SetID(id int)    { w.id = id }

func (w *Weapon) Use() (string, error) {
	if w.Durability <= 0 {
		return "", ErrBroken
	}
	w.Durability--
	return fmt.Sprintf("Используется %s (урон %d, прочность осталась %d)", w.name, w.Damage, w.Durability), nil
}

type Armor struct {
	id       int
	name     string
	weight   float64
	Defense  int
	Equipped bool
}

func (a *Armor) Name() string    { return a.name }
func (a *Armor) Weight() float64 { return a.weight }
func (a *Armor) ID() int         { return a.id }
func (a *Armor) SetID(id int)    { a.id = id }

func (a *Armor) Use() (string, error) {
	if a.Equipped {
		return fmt.Sprintf("%s уже надето", a.name), nil
	}
	a.Equipped = true
	return fmt.Sprintf("Надевается %s (защита %d)", a.name, a.Defense), nil
}

type Potion struct {
	id      int
	name    string
	weight  float64
	Effect  string
	Charges int
}

func (p *Potion) Name() string    { return p.name }
func (p *Potion) Weight() float64 { return p.weight }
func (p *Potion) ID() int         { return p.id }
func (p *Potion) SetID(id int)    { p.id = id }

func (p *Potion) Use() (string, error) {
	if p.Charges <= 0 {
		return "", ErrExhausted
	}
	p.Charges--
	return fmt.Sprintf("Выпито %s (%s), зарядов осталось %d", p.name, p.Effect, p.Charges), nil
}

func DescribeItem(i Item) string {
	if i == nil {
		return "Предмет отсутствует"
	}
	switch v := i.(type) {
	case *Weapon:
		if v == nil {
			return "Предмет отсутствует"
		}
	case *Armor:
		if v == nil {
			return "Предмет отсутствует"
		}
	case *Potion:
		if v == nil {
			return "Предмет отсутствует"
		}
	}
	return fmt.Sprintf("%s (вес: %g)", i.Name(), i.Weight())
}

func Filter[T any](items []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, item := range items {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

func Map[T any, R any](items []T, transform func(T) R) []R {
	result := make([]R, len(items))
	for i, item := range items {
		result[i] = transform(item)
	}
	return result
}

func Find[T any](items []T, condition func(T) bool) (T, bool) {
	var zero T
	for _, item := range items {
		if condition(item) {
			return item, true
		}
	}
	return zero, false
}

type Inventory struct {
	Items  []Item
	nextID int
}

func NewInventory() *Inventory {
	return &Inventory{nextID: 1}
}

func (inv *Inventory) AddItem(item Item) int {
	id := inv.nextID
	inv.nextID++
	item.SetID(id)
	inv.Items = append(inv.Items, item)
	return id
}

func (inv *Inventory) Weapons() []*Weapon {
	weapons := Filter(inv.Items, func(i Item) bool {
		_, ok := i.(*Weapon)
		return ok
	})
	return Map(weapons, func(i Item) *Weapon {
		return i.(*Weapon)
	})
}

func (inv *Inventory) BrokenItems() []Item {
	return Filter(inv.Items, func(i Item) bool {
		switch v := i.(type) {
		case *Weapon:
			return v.Durability <= 0
		case *Potion:
			return v.Charges <= 0
		default:
			return false
		}
	})
}

func (inv *Inventory) ItemNames() []string {
	return Map(inv.Items, func(i Item) string {
		return i.Name()
	})
}

func (inv *Inventory) FindItemByName(name string) (Item, bool) {
	return Find(inv.Items, func(i Item) bool {
		return i.Name() == name
	})
}

func (inv *Inventory) FindItemByID(id int) (Item, bool) {
	return Find(inv.Items, func(i Item) bool {
		return i.ID() == id
	})
}

func (inv *Inventory) DeleteItem(id int) error {
	for idx, item := range inv.Items {
		if item.ID() == id {
			inv.Items = append(inv.Items[:idx], inv.Items[idx+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func serializeItem(item Item) string {
	switch v := item.(type) {
	case *Weapon:
		return fmt.Sprintf("Weapon|%d|%s|%d|%d|%g", v.id, v.name, v.Damage, v.Durability, v.weight)
	case *Armor:
		eq := 0
		if v.Equipped {
			eq = 1
		}
		return fmt.Sprintf("Armor|%d|%s|%d|%g|%d", v.id, v.name, v.Defense, v.weight, eq)
	case *Potion:
		return fmt.Sprintf("Potion|%d|%s|%s|%d|%g", v.id, v.name, v.Effect, v.Charges, v.weight)
	default:
		return ""
	}
}

func parseItem(line string, lineNum int) (Item, error) {
	parts := strings.Split(line, "|")
	if len(parts) < 2 {
		return nil, &ParseItemError{Line: lineNum, Value: line, Err: errors.New("too few fields")}
	}

	kind := parts[0]
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, &ParseItemError{Line: lineNum, Value: line, Err: fmt.Errorf("invalid id: %w", err)}
	}

	switch kind {
	case "Weapon":
		if len(parts) != 6 {
			return nil, &ParseItemError{Line: lineNum, Value: line, Err: errors.New("expected 6 fields for Weapon")}
		}
		dmg, err1 := strconv.Atoi(parts[3])
		dur, err2 := strconv.Atoi(parts[4])
		w, err3 := strconv.ParseFloat(parts[5], 64)
		if err1 != nil || err2 != nil || err3 != nil {
			return nil, &ParseItemError{Line: lineNum, Value: line, Err: errors.New("invalid numeric fields")}
		}
		return &Weapon{id: id, name: parts[2], Damage: dmg, Durability: dur, weight: w}, nil

	case "Armor":
		if len(parts) < 5 || len(parts) > 6 {
			return nil, &ParseItemError{Line: lineNum, Value: line, Err: errors.New("expected 5 or 6 fields for Armor")}
		}
		def, err1 := strconv.Atoi(parts[3])
		w, err2 := strconv.ParseFloat(parts[4], 64)
		if err1 != nil || err2 != nil {
			return nil, &ParseItemError{Line: lineNum, Value: line, Err: errors.New("invalid numeric fields")}
		}
		eq := false
		if len(parts) == 6 {
			if parts[5] == "1" {
				eq = true
			}
		}
		return &Armor{id: id, name: parts[2], Defense: def, weight: w, Equipped: eq}, nil

	case "Potion":
		if len(parts) != 6 {
			return nil, &ParseItemError{Line: lineNum, Value: line, Err: errors.New("expected 6 fields for Potion")}
		}
		ch, err1 := strconv.Atoi(parts[4])
		w, err2 := strconv.ParseFloat(parts[5], 64)
		if err1 != nil || err2 != nil {
			return nil, &ParseItemError{Line: lineNum, Value: line, Err: errors.New("invalid numeric fields")}
		}
		return &Potion{id: id, name: parts[2], Effect: parts[3], Charges: ch, weight: w}, nil

	default:
		return nil, &ParseItemError{Line: lineNum, Value: line, Err: fmt.Errorf("unknown kind %q", kind)}
	}
}

func (inv *Inventory) Save(w io.Writer) error {
	bw := bufio.NewWriter(w)
	for _, item := range inv.Items {
		line := serializeItem(item)
		if line == "" {
			continue
		}
		if _, err := bw.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("write item: %w", err)
		}
	}
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}
	return nil
}

func (inv *Inventory) Load(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	lineNum := 0
	var maxID int

	var tempItems []Item
	idMap := make(map[int]bool)

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		item, err := parseItem(line, lineNum)
		if err != nil {
			return err
		}

		if idMap[item.ID()] {
			return &ParseItemError{Line: lineNum, Value: line, Err: fmt.Errorf("duplicate item id: %d", item.ID())}
		}
		idMap[item.ID()] = true

		tempItems = append(tempItems, item)
		if item.ID() > maxID {
			maxID = item.ID()
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner: %w", err)
	}

	inv.Items = tempItems
	inv.nextID = maxID + 1
	return nil
}

func (inv *Inventory) SaveToFile(path string) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer f.Close()

	if err := inv.Save(f); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

func (inv *Inventory) LoadFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close()
	return inv.Load(f)
}

func UseDangerousItem(item Item) (message string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	if item == nil {
		return "", errors.New("nil item")
	}

	if item.Name() == "Бомба" {
		panic("взрыв!")
	}

	return item.Use()
}

const inventoryFile = "inventory.dat"
const demoInventoryFile = "demo_inventory.dat" // Исправление 5: отдельный файл для демо

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: inventory <command> [flags]")
		return 2
	}

	cmd := args[0]
	rest := args[1:]

	inv := NewInventory()
	if err := inv.LoadFromFile(inventoryFile); err != nil {
		fmt.Fprintf(stderr, "load inventory: %v\n", err)
		var pe *ParseItemError
		if errors.As(err, &pe) {
			fmt.Fprintf(stderr, "  line %d: %q\n", pe.Line, pe.Value)
		}
		return 1
	}

	switch cmd {
	case "list":
		return cmdList(inv, rest, stdout, stderr)
	case "add":
		return cmdAdd(inv, rest, stdout, stderr)
	case "use":
		return cmdUse(inv, rest, stdout, stderr)
	case "delete":
		return cmdDelete(inv, rest, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n", cmd)
		return 2
	}
}

func cmdList(inv *Inventory, args []string, stdout, stderr io.Writer) int {
	fs := flagNew("list")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "list command takes no positional arguments, got %v\n", fs.Args())
		return 2
	}

	if len(inv.Items) == 0 {
		fmt.Fprintln(stdout, "inventory is empty")
		return 0
	}
	for _, item := range inv.Items {
		fmt.Fprintf(stdout, "id=%d  %s\n", item.ID(), DescribeItem(item))
	}
	return 0
}

func cmdAdd(inv *Inventory, args []string, stdout, stderr io.Writer) int {
	fs := flagNew("add")
	name := fs.String("name", "", "item name")
	weight := fs.Float64("weight", 0, "item weight")
	kind := fs.String("type", "weapon", "item type: weapon|armor|potion")
	damage := fs.Int("damage", 5, "weapon damage")
	durability := fs.Int("durability", 10, "weapon durability")
	defense := fs.Int("defense", 5, "armor defense")
	effect := fs.String("effect", "+10 HP", "potion effect")
	charges := fs.Int("charges", 3, "potion charges")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	// Исправление 4: проверка NArg для выявления лишних аргументов
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "unexpected positional arguments for add: %v\n", fs.Args())
		return 2
	}

	if *name == "" {
		fmt.Fprintln(stderr, "flag -name is required")
		return 2
	}

	if strings.Contains(*name, "|") || strings.Contains(*name, "\n") {
		fmt.Fprintln(stderr, "invalid item name: must not contain '|' or newlines")
		return 2
	}
	if strings.Contains(*effect, "|") || strings.Contains(*effect, "\n") {
		fmt.Fprintln(stderr, "invalid potion effect: must not contain '|' or newlines")
		return 2
	}

	var item Item
	switch strings.ToLower(*kind) {
	case "weapon":
		item = &Weapon{name: *name, weight: *weight, Damage: *damage, Durability: *durability}
	case "armor":
		item = &Armor{name: *name, weight: *weight, Defense: *defense}
	case "potion":
		item = &Potion{name: *name, weight: *weight, Effect: *effect, Charges: *charges}
	default:
		fmt.Fprintf(stderr, "unknown type %q (use weapon|armor|potion)\n", *kind)
		return 2
	}

	id := inv.AddItem(item)
	if err := inv.SaveToFile(inventoryFile); err != nil {
		fmt.Fprintf(stderr, "save: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "added item id=%d\n", id)
	return 0
}

func cmdUse(inv *Inventory, args []string, stdout, stderr io.Writer) int {
	fs := flagNew("use")
	id := fs.Int("id", 0, "item id")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "unexpected positional arguments for use: %v\n", fs.Args())
		return 2
	}

	if *id == 0 {
		fmt.Fprintln(stderr, "flag -id is required")
		return 2
	}

	item, ok := inv.FindItemByID(*id)
	if !ok {
		fmt.Fprintf(stderr, "item id=%d not found\n", *id)
		return 1
	}

	msg, err := item.Use()
	if err != nil {
		if errors.Is(err, ErrBroken) {
			fmt.Fprintf(stderr, "cannot use: item is broken\n")
		} else if errors.Is(err, ErrExhausted) {
			fmt.Fprintf(stderr, "cannot use: item is exhausted\n")
		} else {
			fmt.Fprintf(stderr, "use error: %v\n", err)
		}
		return 1
	}

	if err := inv.SaveToFile(inventoryFile); err != nil {
		fmt.Fprintf(stderr, "save: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, msg)
	return 0
}

func cmdDelete(inv *Inventory, args []string, stdout, stderr io.Writer) int {
	fs := flagNew("delete")
	id := fs.Int("id", 0, "item id")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "unexpected positional arguments for delete: %v\n", fs.Args())
		return 2
	}

	if *id == 0 {
		fmt.Fprintln(stderr, "flag -id is required")
		return 2
	}

	if err := inv.DeleteItem(*id); err != nil {
		fmt.Fprintf(stderr, "delete: %v\n", err)
		return 1
	}
	if err := inv.SaveToFile(inventoryFile); err != nil {
		fmt.Fprintf(stderr, "save: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "deleted item id=%d\n", *id)
	return 0
}

func flagNew(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "inventory" {
		args = args[1:]
	}

	if len(args) == 0 {
		demo()
		return
	}

	code := run(args, os.Stdout, os.Stderr)
	os.Exit(code)
}

func demo() {
	inv := NewInventory()

	sword := &Weapon{name: "Меч", weight: 3.0, Damage: 10, Durability: 5}
	shield := &Armor{name: "Щит", weight: 4.5, Defense: 5}
	potion := &Potion{name: "Лечебное", weight: 0.5, Effect: "+50 HP", Charges: 3}
	bow := &Weapon{name: "Сломанный лук", weight: 1.5, Damage: 5, Durability: 0}

	inv.AddItem(sword)
	inv.AddItem(shield)
	inv.AddItem(potion)
	inv.AddItem(bow)

	fmt.Println("=== Use ===")
	if msg, err := sword.Use(); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(msg)
	}
	if msg, err := bow.Use(); err != nil {
		if errors.Is(err, ErrBroken) {
			fmt.Println("broken:", err)
		} else {
			fmt.Println("error:", err)
		}
	} else {
		fmt.Println(msg)
	}

	fmt.Println("\n=== DescribeItem ===")
	fmt.Println(DescribeItem(sword))
	fmt.Println(DescribeItem(nil))
	var nilWeapon *Weapon
	fmt.Println(DescribeItem(nilWeapon))

	fmt.Println("\n=== Weapons ===")
	for _, w := range inv.Weapons() {
		fmt.Printf("  %s (dur=%d)\n", w.Name(), w.Durability)
	}

	fmt.Println("\n=== BrokenItems ===")
	for _, i := range inv.BrokenItems() {
		fmt.Printf("  %s\n", i.Name())
	}

	fmt.Println("\n=== ItemNames ===")
	fmt.Println(inv.ItemNames())

	fmt.Println("\n=== FindItemByName(\"Щит\") ===")
	if item, ok := inv.FindItemByName("Щит"); ok {
		fmt.Println(DescribeItem(item))
	} else {
		fmt.Println("not found")
	}

	fmt.Println("\n=== UseDangerousItem ===")
	bomb := &Weapon{name: "Бомба", weight: 1, Damage: 100, Durability: 1}
	if msg, err := UseDangerousItem(bomb); err != nil {
		fmt.Println("recovered:", err)
	} else {
		fmt.Println(msg)
	}

	fmt.Println("\n=== Save / Load ===")
	if err := inv.SaveToFile(demoInventoryFile); err != nil {
		fmt.Println("save error:", err)
	} else {
		fmt.Println("saved to", demoInventoryFile)
	}

	inv2 := NewInventory()
	if err := inv2.LoadFromFile(demoInventoryFile); err != nil {
		fmt.Println("load error:", err)
		var pe *ParseItemError
		if errors.As(err, &pe) {
			fmt.Printf("  line %d: %q\n", pe.Line, pe.Value)
		}
	} else {
		fmt.Println("loaded names:", inv2.ItemNames())
	}
}
