package main

import "fmt"

type Profile struct {
	Name       string
	Age        int
	Email      string
	Newsletter bool
}

// ProfilePatch описывает, какие поля профиля нужно изменить.
// Поле nil значит «не менять», а не-nil значит «записать это значение»,
// в том числе 0 или false.
type ProfilePatch struct {
	Name       *string
	Age        *int
	Email      *string
	Newsletter *bool
}

// ApplyPatch применяет только те поля патча, которые не nil.
// Если профиль nil, ничего не делает: без этой проверки обращение
// к p.Name вызвало бы панику.
func ApplyPatch(p *Profile, patch ProfilePatch) {
	if p == nil {
		return
	}
	if patch.Name != nil {
		p.Name = *patch.Name
	}
	if patch.Age != nil {
		p.Age = *patch.Age
	}
	if patch.Email != nil {
		p.Email = *patch.Email
	}
	if patch.Newsletter != nil {
		p.Newsletter = *patch.Newsletter
	}
}

// Ответ. У bool всего два значения, true и false, и «поле не задано» среди
// них нет. Если в патче поле Newsletter bool, то значение по умолчанию false
// невозможно отличить от осознанного «выключить подписку»: ApplyPatch не
// понял бы, нужно ли записать false или поле вообще не трогать. То же с
// Age int и значением 0, и с пустой строкой в Name.
//
// У указателя *bool три состояния:
//   - nil: поле не передали, профиль не меняем;
//   - указатель на true: записать true;
//   - указатель на false: записать false.
//
// Так «не менять» и «записать ноль» различаются.

func main() {
	profile := Profile{Name: "Анна", Age: 30, Email: "anna@example.com", Newsletter: true}
	fmt.Printf("%+v\n", profile)

	// Литерал нельзя передать как указатель (&31 не компилируется),
	// поэтому значения сначала кладём в переменные и берём их адреса.
	age := 31
	newsletter := false
	ApplyPatch(&profile, ProfilePatch{Age: &age, Newsletter: &newsletter})
	fmt.Printf("%+v\n", profile)

	// Пустой патч профиль не меняет: все поля nil.
	ApplyPatch(&profile, ProfilePatch{})
	fmt.Printf("%+v\n", profile)

	// Записать 0 тоже можно: указатель не nil, поэтому значение применится.
	zero := 0
	ApplyPatch(&profile, ProfilePatch{Age: &zero})
	fmt.Printf("%+v\n", profile)

	// nil-профиль не вызывает панику.
	name := "Борис"
	ApplyPatch(nil, ProfilePatch{Name: &name})
	fmt.Println("ApplyPatch(nil, ...) отработал без паники")
}
