package storage

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"golang_lessons/lesson_7/game/inventory"
)

// Помощники и тестовые реализации для всех тестовых файлов пакета.

func mustWeapon(t *testing.T, name string, weight float64, damage, durability int) *inventory.Weapon {
	t.Helper()
	w, err := inventory.NewWeapon(name, weight, damage, durability)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func mustArmor(t *testing.T, name string, weight float64, defense int, equipped bool) *inventory.Armor {
	t.Helper()
	a, err := inventory.NewArmor(name, weight, defense)
	if err != nil {
		t.Fatal(err)
	}
	a.Equipped = equipped
	return a
}

func mustPotion(t *testing.T, name string, weight float64, effect string, charges int) *inventory.Potion {
	t.Helper()
	p, err := inventory.NewPotion(name, weight, effect, charges)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// newInventory создаёт инвентарь из предметов; ID выдаются подряд, начиная с 1.
func newInventory(t *testing.T, items ...inventory.Item) *inventory.Inventory {
	t.Helper()
	inv := inventory.New()
	for _, item := range items {
		if _, err := inv.AddItem(item); err != nil {
			t.Fatal(err)
		}
	}
	return inv
}

// sampleInventory — меч, надетый щит и лечебное зелье с ID 1, 2 и 3.
func sampleInventory(t *testing.T) *inventory.Inventory {
	return newInventory(t,
		mustWeapon(t, "Меч", 3.5, 10, 5),
		mustArmor(t, "Щит", 4.5, 5, true),
		mustPotion(t, "Лечебное", 0.3, "+50 HP", 3),
	)
}

// sampleDump — sampleInventory в формате хранения.
const sampleDump = "Weapon|1|Меч|10|5|3.5\nArmor|2|Щит|5|4.5|1\nPotion|3|Лечебное|+50 HP|3|0.3\n"

// dump превращает инвентарь в текст формата хранения: так удобно сравнивать содержимое.
func dump(t *testing.T, inv *inventory.Inventory) string {
	t.Helper()
	var buf bytes.Buffer
	if err := encode(&buf, inv); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

// fakeItem — предмет типа, который хранилище не умеет сохранять.
type fakeItem struct{}

func (fakeItem) Name() string         { return "Пустышка" }
func (fakeItem) Weight() float64      { return 1 }
func (fakeItem) Use() (string, error) { return "", nil }

// failingWriter принимает не больше failAfter байт, а затем возвращает ошибку err.
type failingWriter struct {
	failAfter int
	err       error
	written   int
}

func (w *failingWriter) Write(p []byte) (int, error) {
	room := w.failAfter - w.written
	if len(p) <= room {
		w.written += len(p)
		return len(p), nil
	}
	if room < 0 {
		room = 0
	}
	w.written += room
	return room, w.err
}

func TestFormatItem(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		item    inventory.Item
		want    string
		wantErr error
	}{
		{"оружие", 1, mustWeapon(t, "Меч", 3.5, 10, 5), "Weapon|1|Меч|10|5|3.5", nil},
		{"оружие с целым весом", 2, mustWeapon(t, "Лук", 3, 5, 0), "Weapon|2|Лук|5|0|3", nil},
		{"оружие с нулевым весом", 3, mustWeapon(t, "Нож", 0, 1, 1), "Weapon|3|Нож|1|1|0", nil},
		{"броня не надета", 4, mustArmor(t, "Щит", 4.5, 5, false), "Armor|4|Щит|5|4.5|0", nil},
		{"броня надета", 5, mustArmor(t, "Щит", 4.5, 5, true), "Armor|5|Щит|5|4.5|1", nil},
		{"зелье", 6, mustPotion(t, "Лечебное", 0.3, "+50 HP", 3), "Potion|6|Лечебное|+50 HP|3|0.3", nil},
		{"зелье с пустым эффектом", 7, mustPotion(t, "Вода", 0.1, "", 1), "Potion|7|Вода||1|0.1", nil},
		{"неизвестный тип предмета", 8, fakeItem{}, "", ErrNotStorable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatItem(tt.id, tt.item)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, хочу %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("строка = %q, хочу %q", got, tt.want)
			}
		})
	}
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantID      int
		wantRound   string // что получится, если снова превратить разобранный предмет в строку
		wantErrIs   error  // errors.Is-проверка при ошибке
		wantErrText string // подстрока в тексте ошибки
		wantFail    bool
	}{
		// Успешный разбор.
		{name: "оружие", line: "Weapon|1|Меч|10|5|3.5", wantID: 1, wantRound: "Weapon|1|Меч|10|5|3.5"},
		{name: "броня надета", line: "Armor|2|Щит|5|4.5|1", wantID: 2, wantRound: "Armor|2|Щит|5|4.5|1"},
		{name: "броня не надета", line: "Armor|2|Щит|5|4.5|0", wantID: 2, wantRound: "Armor|2|Щит|5|4.5|0"},
		{name: "старая запись брони без Equipped", line: "Armor|2|Щит|5|4.5", wantID: 2, wantRound: "Armor|2|Щит|5|4.5|0"},
		{name: "зелье", line: "Potion|3|Лечебное|+50 HP|3|0.3", wantID: 3, wantRound: "Potion|3|Лечебное|+50 HP|3|0.3"},
		{name: "отрицательные характеристики", line: "Weapon|7|Меч|-1|-2|0", wantID: 7, wantRound: "Weapon|7|Меч|-1|-2|0"},
		{name: "вес в экспоненциальной записи", line: "Weapon|1|Меч|1|1|1e2", wantID: 1, wantRound: "Weapon|1|Меч|1|1|100"},
		{name: "ID не проверяется на положительность здесь", line: "Weapon|0|Меч|1|1|1", wantID: 0, wantRound: "Weapon|0|Меч|1|1|1"},

		// Общие ошибки.
		{name: "одно поле", line: "Weapon", wantFail: true, wantErrText: "too few fields"},
		{name: "ID не число", line: "Weapon|abc|Меч|1|1|1", wantFail: true, wantErrText: "invalid id"},
		{name: "неизвестный тип", line: "Gun|1|Меч|1|1|1", wantFail: true, wantErrText: "unknown kind"},

		// Оружие.
		{name: "оружие: мало полей", line: "Weapon|1|Меч|1|1", wantFail: true, wantErrText: "expected 6 fields for Weapon"},
		{name: "оружие: много полей", line: "Weapon|1|Меч|1|1|1|x", wantFail: true, wantErrText: "expected 6 fields for Weapon"},
		{name: "оружие: урон не число", line: "Weapon|1|Меч|x|1|1", wantFail: true, wantErrText: "invalid damage"},
		{name: "оружие: прочность не число", line: "Weapon|1|Меч|1|x|1", wantFail: true, wantErrText: "invalid durability"},
		{name: "оружие: вес не число", line: "Weapon|1|Меч|1|1|heavy", wantFail: true, wantErrText: "invalid weight"},
		{name: "оружие: пустое название", line: "Weapon|1||1|1|1", wantFail: true, wantErrIs: inventory.ErrInvalidItem},
		{name: "оружие: отрицательный вес", line: "Weapon|1|Меч|1|1|-1", wantFail: true, wantErrIs: inventory.ErrInvalidItem},
		{name: "оружие: NaN", line: "Weapon|1|Меч|1|1|NaN", wantFail: true, wantErrIs: inventory.ErrInvalidItem},
		{name: "оружие: Inf", line: "Weapon|1|Меч|1|1|Inf", wantFail: true, wantErrIs: inventory.ErrInvalidItem},

		// Броня.
		{name: "броня: мало полей", line: "Armor|1|Щит|5", wantFail: true, wantErrText: "expected 5 or 6 fields for Armor"},
		{name: "броня: много полей", line: "Armor|1|Щит|5|4.5|1|2", wantFail: true, wantErrText: "expected 5 or 6 fields for Armor"},
		{name: "броня: защита не число", line: "Armor|1|Щит|x|4.5|0", wantFail: true, wantErrText: "invalid defense"},
		{name: "броня: вес не число", line: "Armor|1|Щит|5|x|0", wantFail: true, wantErrText: "invalid weight"},
		{name: "броня: неверный флаг", line: "Armor|1|Щит|5|4.5|2", wantFail: true, wantErrText: "invalid equipped flag"},
		{name: "броня: пустое название", line: "Armor|1||5|4.5|0", wantFail: true, wantErrIs: inventory.ErrInvalidItem},

		// Зелье.
		{name: "зелье: мало полей", line: "Potion|1|Зелье|эффект", wantFail: true, wantErrText: "expected 6 fields for Potion"},
		{name: "зелье: заряды не число", line: "Potion|1|Зелье|эффект|x|1", wantFail: true, wantErrText: "invalid charges"},
		{name: "зелье: вес не число", line: "Potion|1|Зелье|эффект|1|x", wantFail: true, wantErrText: "invalid weight"},
		{name: "зелье: пустое название", line: "Potion|1||эффект|1|1", wantFail: true, wantErrIs: inventory.ErrInvalidItem},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, item, err := parseLine(tt.line)

			if tt.wantFail {
				if err == nil {
					t.Fatalf("ошибки нет, разобрано %v", item)
				}
				if item != nil {
					t.Errorf("вместе с ошибкой вернулся предмет %v", item)
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("err = %v, хочу %v", err, tt.wantErrIs)
				}
				if tt.wantErrText != "" && !strings.Contains(err.Error(), tt.wantErrText) {
					t.Errorf("текст ошибки %q не содержит %q", err, tt.wantErrText)
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}
			if id != tt.wantID {
				t.Errorf("ID = %d, хочу %d", id, tt.wantID)
			}
			got, formatErr := formatItem(id, item)
			if formatErr != nil {
				t.Fatal(formatErr)
			}
			if got != tt.wantRound {
				t.Errorf("после разбора и форматирования: %q, хочу %q", got, tt.wantRound)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	withoutSecond := sampleInventory(t)
	if err := withoutSecond.Delete(2); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		inv  *inventory.Inventory
		want string
	}{
		{"пустой инвентарь", inventory.New(), ""},
		{"все три типа", sampleInventory(t), sampleDump},
		{"ID после удаления сохраняются", withoutSecond, "Weapon|1|Меч|10|5|3.5\nPotion|3|Лечебное|+50 HP|3|0.3\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := encode(&buf, tt.inv); err != nil {
				t.Fatal(err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("encode = %q, хочу %q", got, tt.want)
			}
		})
	}
}

func TestEncodeErrors(t *testing.T) {
	boom := errors.New("диск отвалился")

	t.Run("неизвестный тип предмета", func(t *testing.T) {
		inv := newInventory(t, mustWeapon(t, "Меч", 3.5, 10, 5), fakeItem{})
		var buf bytes.Buffer

		err := encode(&buf, inv)

		if !errors.Is(err, ErrNotStorable) {
			t.Fatalf("err = %v, хочу ErrNotStorable", err)
		}
		if !strings.Contains(err.Error(), "id=2") {
			t.Errorf("в ошибке нет ID проблемного предмета: %q", err)
		}
	})

	writerTests := []struct {
		name      string
		failAfter int
	}{
		{"писатель падает сразу", 0},
		{"писатель падает посреди строки", 10},
		{"писатель падает после первой строки", len("Weapon|1|Меч|10|5|3.5\n")},
	}
	for _, tt := range writerTests {
		t.Run(tt.name, func(t *testing.T) {
			w := &failingWriter{failAfter: tt.failAfter, err: boom}

			err := encode(w, sampleInventory(t))

			if !errors.Is(err, boom) {
				t.Errorf("err = %v, хочу исходную ошибку записи", err)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string // результат в формате хранения (для успешных случаев)
		wantLine  int    // номер строки в ParseItemError; 0 — ошибки быть не должно
		wantValue string // содержимое строки в ParseItemError
		wantIs    error  // errors.Is-проверка причины
	}{
		{name: "пустой ввод", input: "", want: ""},
		{name: "только пустые строки", input: "\n\n  \n\t\n", want: ""},
		{name: "обычные данные", input: sampleDump, want: sampleDump},
		{
			name:  "строки с CRLF",
			input: "Weapon|1|Меч|10|5|3.5\r\nArmor|2|Щит|5|4.5|1\r\n",
			want:  "Weapon|1|Меч|10|5|3.5\nArmor|2|Щит|5|4.5|1\n",
		},
		{name: "нет перевода строки в конце", input: "Weapon|1|Меч|10|5|3.5", want: "Weapon|1|Меч|10|5|3.5\n"},
		{name: "пробелы вокруг строки", input: "  Weapon|1|Меч|10|5|3.5  \n", want: "Weapon|1|Меч|10|5|3.5\n"},
		{
			name:  "пустые строки между записями",
			input: "Weapon|1|Меч|10|5|3.5\n\nArmor|2|Щит|5|4.5|0\n",
			want:  "Weapon|1|Меч|10|5|3.5\nArmor|2|Щит|5|4.5|0\n",
		},
		{
			name:  "ID не подряд и порядок как в файле",
			input: "Weapon|5|Меч|10|5|3.5\nWeapon|2|Лук|5|0|1.2\n",
			want:  "Weapon|5|Меч|10|5|3.5\nWeapon|2|Лук|5|0|1.2\n",
		},
		{
			name: "повреждена первая строка", input: "бред\nWeapon|1|Меч|10|5|3.5\n",
			wantLine: 1, wantValue: "бред",
		},
		{
			name:     "номер строки считает и пустые строки",
			input:    "Weapon|1|Меч|10|5|3.5\n\nбред\n",
			wantLine: 3, wantValue: "бред",
		},
		{
			name:     "повреждена строка с неверным числом",
			input:    "Weapon|1|Меч|10|5|3.5\nArmor|2|Щит|abc|4.5|0\nPotion|3|Зелье|эффект|1|1\n",
			wantLine: 2, wantValue: "Armor|2|Щит|abc|4.5|0",
		},
		{
			name:     "повторяющийся ID",
			input:    "Weapon|1|Меч|10|5|3.5\nWeapon|1|Лук|5|0|1.2\n",
			wantLine: 2, wantValue: "Weapon|1|Лук|5|0|1.2", wantIs: inventory.ErrDuplicateID,
		},
		{
			name:     "нулевой ID",
			input:    "Weapon|0|Меч|1|1|1\n",
			wantLine: 1, wantValue: "Weapon|0|Меч|1|1|1", wantIs: inventory.ErrInvalidID,
		},
		{
			name:     "отрицательный ID",
			input:    "Weapon|-3|Меч|1|1|1\n",
			wantLine: 1, wantValue: "Weapon|-3|Меч|1|1|1", wantIs: inventory.ErrInvalidID,
		},
		{
			name:     "неверное значение отклоняется правилами конструктора",
			input:    "Weapon|1|Меч|1|1|-1\n",
			wantLine: 1, wantValue: "Weapon|1|Меч|1|1|-1", wantIs: inventory.ErrInvalidItem,
		},
		{
			name:     "значение в ошибке без пробелов по краям",
			input:    "  бред  \n",
			wantLine: 1, wantValue: "бред",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, err := decode(strings.NewReader(tt.input))

			if tt.wantLine == 0 {
				if err != nil {
					t.Fatal(err)
				}
				if got := dump(t, inv); got != tt.want {
					t.Errorf("загружено %q, хочу %q", got, tt.want)
				}
				return
			}

			if inv != nil {
				t.Error("при ошибке инвентарь не должен возвращаться")
			}
			var parseErr *ParseItemError
			if !errors.As(err, &parseErr) {
				t.Fatalf("err = %v, хочу *ParseItemError", err)
			}
			if parseErr.Line != tt.wantLine || parseErr.Value != tt.wantValue {
				t.Errorf("Line = %d, Value = %q; хочу %d, %q", parseErr.Line, parseErr.Value, tt.wantLine, tt.wantValue)
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Errorf("err = %v, хочу %v", err, tt.wantIs)
			}
		})
	}
}

func TestDecodeReaderErrors(t *testing.T) {
	boom := errors.New("диск отвалился")

	t.Run("ошибка чтения не маскируется под повреждение данных", func(t *testing.T) {
		r := io.MultiReader(strings.NewReader("Weapon|1|Меч|10|5|3.5\n"), iotest.ErrReader(boom))

		inv, err := decode(r)

		if !errors.Is(err, boom) {
			t.Fatalf("err = %v, хочу исходную ошибку чтения", err)
		}
		var parseErr *ParseItemError
		if errors.As(err, &parseErr) {
			t.Errorf("ошибка чтения не должна быть ParseItemError: %v", err)
		}
		if inv != nil {
			t.Error("при ошибке инвентарь не должен возвращаться")
		}
	})

	t.Run("длинная строка читается (нет лимита Scanner)", func(t *testing.T) {
		// bufio.Reader не ограничивает длину строки 64 KiB.
		// Строка из одних 'a' — невалидный формат → ParseItemError, не ErrTooLong.
		long := strings.Repeat("a", bufio.MaxScanTokenSize+1)

		inv, err := decode(strings.NewReader(long))

		if errors.Is(err, bufio.ErrTooLong) {
			t.Fatalf("не ожидаем bufio.ErrTooLong при чтении через Reader: %v", err)
		}
		var parseErr *ParseItemError
		if !errors.As(err, &parseErr) {
			t.Fatalf("err = %v, хочу *ParseItemError", err)
		}
		if inv != nil {
			t.Error("при ошибке инвентарь не должен возвращаться")
		}
	})

	t.Run("успешно сохранённая длинная запись читается обратно", func(t *testing.T) {
		// Имя длиннее MaxScanTokenSize
		name := strings.Repeat("x", bufio.MaxScanTokenSize+1)
		potion := mustPotion(t, name, 0.5, "+HP", 1)
		inv := inventory.New()
		if _, err := inv.AddItem(potion); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if err := encode(&buf, inv); err != nil {
			t.Fatal(err)
		}
		loaded, err := decode(&buf)
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got := loaded.ItemNames(); len(got) != 1 || got[0] != name {
			t.Errorf("names = %v, хочу [%q]", got, name)
		}
	})
}

func TestParseItemError(t *testing.T) {
	cause := errors.New("причина")
	err := &ParseItemError{Line: 7, Value: "bad|line", Err: cause}

	if !errors.Is(err, cause) {
		t.Error("errors.Is не видит причину через Unwrap")
	}
	if err.Unwrap() != cause {
		t.Errorf("Unwrap = %v, хочу %v", err.Unwrap(), cause)
	}
	text := err.Error()
	for _, want := range []string{"line 7", `"bad|line"`, "причина"} {
		if !strings.Contains(text, want) {
			t.Errorf("текст ошибки %q не содержит %q", text, want)
		}
	}

	// Ошибку можно достать из цепочки обёрток: так её видит верхний уровень.
	wrapped := errors.Join(errors.New("другая ошибка"), err)
	var found *ParseItemError
	if !errors.As(wrapped, &found) || found.Line != 7 {
		t.Errorf("errors.As не нашёл ParseItemError в цепочке: %v", wrapped)
	}
}

// Effect экспортируемый: его можно изменить после NewPotion.
// encode должен отклонить «битое» значение, иначе файл нельзя будет прочитать.
func TestEncodeRejectsMutatedEffect(t *testing.T) {
	potion := mustPotion(t, "Зелье", 0.5, "+10 HP", 1)
	potion.Effect = "a|b" // обход конструктора

	inv := inventory.New()
	if _, err := inv.AddItem(potion); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := encode(&buf, inv)
	if !errors.Is(err, inventory.ErrInvalidItem) {
		t.Errorf("err = %v, хочу ErrInvalidItem", err)
	}
	if buf.Len() != 0 {
		t.Errorf("при ошибке в буфер ничего не должно попасть, записано %d байт", buf.Len())
	}
}
