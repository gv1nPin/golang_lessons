// Package storage сохраняет инвентарь и загружает его обратно.
//
// Формат: одна строка на предмет, первое поле — тип, второе — ID:
//
//	Weapon|ID|Name|Damage|Durability|Weight
//	Armor|ID|Name|Defense|Weight|Equipped
//	Potion|ID|Name|Effect|Charges|Weight
//
// Equipped у брони — 1 или 0; старые записи брони без этого поля тоже читаются.
//
// Два хранилища, FileStore и MemoryStore, используют один и тот же формат и одни и те же
// правила (пустое хранилище — пустой инвентарь, ошибка записи не портит прежние данные,
// повреждённая строка даёт *ParseItemError), поэтому взаимозаменяемы.
package storage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang_lessons/lesson_6/game/inventory"
)

// Типы записей и разделитель полей.
const (
	kindWeapon = "Weapon"
	kindArmor  = "Armor"
	kindPotion = "Potion"

	separator = "|"
)

// ErrNotStorable возвращает сохранение, когда в инвентаре есть предмет неизвестного типа.
// Молча пропускать такой предмет нельзя: после перезапуска он бы пропал.
var ErrNotStorable = errors.New("item is not storable")

// ParseItemError описывает повреждённую строку данных.
type ParseItemError struct {
	Line  int    // номер строки, начиная с 1
	Value string // содержимое строки
	Err   error  // причина, по которой строку не удалось разобрать
}

// Error возвращает текст ошибки с номером и содержимым строки.
func (e *ParseItemError) Error() string {
	return fmt.Sprintf("parse error on line %d %q: %v", e.Line, e.Value, e.Err)
}

// Unwrap возвращает причину ошибки, чтобы работали errors.Is и errors.As.
func (e *ParseItemError) Unwrap() error {
	return e.Err
}

// encode записывает инвентарь по одной строке на предмет.
func encode(w io.Writer, inv *inventory.Inventory) error {
	for _, entry := range inv.Entries() {
		line, err := formatItem(entry.ID, entry.Item)
		if err != nil {
			return fmt.Errorf("item id=%d: %w", entry.ID, err)
		}
		if _, err := io.WriteString(w, line+"\n"); err != nil {
			return fmt.Errorf("write item id=%d: %w", entry.ID, err)
		}
	}
	return nil
}

// formatItem превращает предмет в строку данных.
func formatItem(id int, item inventory.Item) (string, error) {
	switch v := item.(type) {
	case *inventory.Weapon:
		return fmt.Sprintf("%s|%d|%s|%d|%d|%g", kindWeapon, id, v.Name(), v.Damage, v.Durability, v.Weight()), nil
	case *inventory.Armor:
		equipped := 0
		if v.Equipped {
			equipped = 1
		}
		return fmt.Sprintf("%s|%d|%s|%d|%g|%d", kindArmor, id, v.Name(), v.Defense, v.Weight(), equipped), nil
	case *inventory.Potion:
		return fmt.Sprintf("%s|%d|%s|%s|%d|%g", kindPotion, id, v.Name(), v.Effect, v.Charges, v.Weight()), nil
	}
	return "", fmt.Errorf("%w: %T", ErrNotStorable, item)
}

// decode читает инвентарь.
//
// Политика загрузки: остановиться на первой повреждённой строке и вернуть *ParseItemError
// с её номером и содержимым (fail-fast). Пропускать строки и возвращать объединённые ошибки
// (errors.Join) опаснее: после «успешной» частичной загрузки следующая команда
// (add, use, delete) сохранила бы неполный инвентарь поверх данных, и повреждённая запись,
// которую можно было поправить вручную, исчезла бы навсегда. Инвентарь собирается во
// временной переменной и возвращается только целиком, поэтому после ошибки частично
// загруженного состояния не остаётся. Дубликат ID считается повреждением строки.
// Пустые строки пропускаются, пустые данные дают пустой инвентарь.
func decode(r io.Reader) (*inventory.Inventory, error) {
	inv := inventory.New()

	scanner := bufio.NewScanner(r)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		id, item, err := parseLine(line)
		if err == nil {
			err = inv.AddItemWithID(id, item)
		}
		if err != nil {
			return nil, &ParseItemError{Line: lineNo, Value: line, Err: err}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	return inv, nil
}

// parseLine разбирает строку данных на ID и предмет.
func parseLine(line string) (int, inventory.Item, error) {
	parts := strings.Split(line, separator)
	if len(parts) < 2 {
		return 0, nil, errors.New("too few fields")
	}
	id, err := parseInt("id", parts[1])
	if err != nil {
		return 0, nil, err
	}

	var item inventory.Item
	switch parts[0] {
	case kindWeapon:
		item, err = parseWeapon(parts)
	case kindArmor:
		item, err = parseArmor(parts)
	case kindPotion:
		item, err = parsePotion(parts)
	default:
		err = fmt.Errorf("unknown kind %q", parts[0])
	}
	if err != nil {
		return 0, nil, err
	}
	return id, item, nil
}

// Предметы создаются через конструкторы inventory, поэтому данные из хранилища проходят
// те же проверки, что и значения из командной строки.

func parseWeapon(parts []string) (inventory.Item, error) {
	if len(parts) != 6 {
		return nil, fmt.Errorf("expected 6 fields for Weapon, got %d", len(parts))
	}
	damage, err := parseInt("damage", parts[3])
	if err != nil {
		return nil, err
	}
	durability, err := parseInt("durability", parts[4])
	if err != nil {
		return nil, err
	}
	weight, err := parseFloat("weight", parts[5])
	if err != nil {
		return nil, err
	}
	weapon, err := inventory.NewWeapon(parts[2], weight, damage, durability)
	if err != nil {
		return nil, err
	}
	return weapon, nil
}

func parseArmor(parts []string) (inventory.Item, error) {
	if len(parts) != 5 && len(parts) != 6 {
		return nil, fmt.Errorf("expected 5 or 6 fields for Armor, got %d", len(parts))
	}
	defense, err := parseInt("defense", parts[3])
	if err != nil {
		return nil, err
	}
	weight, err := parseFloat("weight", parts[4])
	if err != nil {
		return nil, err
	}
	armor, err := inventory.NewArmor(parts[2], weight, defense)
	if err != nil {
		return nil, err
	}
	if len(parts) == 6 {
		switch parts[5] {
		case "1":
			armor.Equipped = true
		case "0":
		default:
			return nil, fmt.Errorf("invalid equipped flag %q: expected 0 or 1", parts[5])
		}
	}
	return armor, nil
}

func parsePotion(parts []string) (inventory.Item, error) {
	if len(parts) != 6 {
		return nil, fmt.Errorf("expected 6 fields for Potion, got %d", len(parts))
	}
	charges, err := parseInt("charges", parts[4])
	if err != nil {
		return nil, err
	}
	weight, err := parseFloat("weight", parts[5])
	if err != nil {
		return nil, err
	}
	potion, err := inventory.NewPotion(parts[2], weight, parts[3], charges)
	if err != nil {
		return nil, err
	}
	return potion, nil
}

// parseInt и parseFloat выражают одно правило: числовое поле с понятным сообщением об ошибке.

func parseInt(field, text string) (int, error) {
	n, err := strconv.Atoi(text)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", field, text, err)
	}
	return n, nil
}

func parseFloat(field, text string) (float64, error) {
	f, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", field, text, err)
	}
	return f, nil
}
