package inventory

import (
	"errors"
	"strings"
	"testing"
)

func TestUseDangerousItem(t *testing.T) {
	errCustom := errors.New("своя ошибка предмета")
	var nilWeapon *Weapon

	tests := []struct {
		name         string
		item         Item
		wantErr      error  // nil: ошибки быть не должно
		wantPanic    bool   // должна ли ошибка быть ErrPanic
		wantInError  string // подстрока в тексте ошибки
		wantInOutput string // подстрока в сообщении при успехе
	}{
		{
			name:        "бомба паникует",
			item:        mustWeapon(t, "Бомба", 1),
			wantErr:     ErrPanic,
			wantPanic:   true,
			wantInError: "взрыв!",
		},
		{
			name:        "паника внутри Use",
			item:        panickyItem{},
			wantErr:     ErrPanic,
			wantPanic:   true,
			wantInError: "ловушка сработала",
		},
		{
			name:    "nil вместо предмета",
			item:    nil,
			wantErr: ErrNilItem,
		},
		{
			name:    "nil-указатель вместо предмета",
			item:    nilWeapon,
			wantErr: ErrNilItem,
		},
		{
			name:    "сломанное оружие остаётся обычной ошибкой",
			item:    mustWeapon(t, "Меч", 0),
			wantErr: ErrBroken,
		},
		{
			name:    "пустое зелье остаётся обычной ошибкой",
			item:    mustPotion(t, "Зелье", 0),
			wantErr: ErrExhausted,
		},
		{
			name:    "чужая ошибка проходит без изменений",
			item:    errorItem{err: errCustom},
			wantErr: errCustom,
		},
		{
			name:         "обычное оружие",
			item:         mustWeapon(t, "Меч", 2),
			wantInOutput: "Меч",
		},
		{
			name:         "броня",
			item:         mustArmor(t, "Щит"),
			wantInOutput: "Надевается Щит",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, err := UseDangerousItem(tt.item)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, хочу %v", err, tt.wantErr)
			}
			if got := errors.Is(err, ErrPanic); got != tt.wantPanic {
				t.Errorf("errors.Is(err, ErrPanic) = %v, хочу %v", got, tt.wantPanic)
			}
			if tt.wantErr != nil {
				if message != "" {
					t.Errorf("при ошибке сообщение должно быть пустым, получено %q", message)
				}
				if tt.wantInError != "" && !strings.Contains(err.Error(), tt.wantInError) {
					t.Errorf("текст ошибки %q не содержит %q", err, tt.wantInError)
				}
				return
			}
			if !strings.Contains(message, tt.wantInOutput) {
				t.Errorf("сообщение %q не содержит %q", message, tt.wantInOutput)
			}
		})
	}
}

// После паники Use не вызывается, поэтому состояние бомбы не меняется;
// у обычных предметов состояние меняется как при обычном Use.
func TestUseDangerousItemState(t *testing.T) {
	t.Run("бомба не изношена", func(t *testing.T) {
		bomb := mustWeapon(t, "Бомба", 3)
		if _, err := UseDangerousItem(bomb); !errors.Is(err, ErrPanic) {
			t.Fatalf("err = %v, хочу ErrPanic", err)
		}
		if bomb.Durability != 3 {
			t.Errorf("прочность = %d, хочу 3", bomb.Durability)
		}
	})

	t.Run("обычное оружие изнашивается", func(t *testing.T) {
		sword := mustWeapon(t, "Меч", 2)
		if _, err := UseDangerousItem(sword); err != nil {
			t.Fatal(err)
		}
		if sword.Durability != 1 {
			t.Errorf("прочность = %d, хочу 1", sword.Durability)
		}
	})

	t.Run("сломанное оружие не меняется", func(t *testing.T) {
		sword := mustWeapon(t, "Меч", 0)
		if _, err := UseDangerousItem(sword); !errors.Is(err, ErrBroken) {
			t.Fatalf("err = %v, хочу ErrBroken", err)
		}
		if sword.Durability != 0 {
			t.Errorf("прочность = %d, хочу 0", sword.Durability)
		}
	})
}

// Паника не вырывается наружу и не мешает следующему вызову.
func TestUseDangerousItemSurvivesPanics(t *testing.T) {
	for i := 0; i < 3; i++ {
		if _, err := UseDangerousItem(panickyItem{}); !errors.Is(err, ErrPanic) {
			t.Fatalf("вызов %d: err = %v, хочу ErrPanic", i, err)
		}
	}
	if _, err := UseDangerousItem(mustWeapon(t, "Меч", 1)); err != nil {
		t.Errorf("обычный вызов после паник: %v", err)
	}
}
