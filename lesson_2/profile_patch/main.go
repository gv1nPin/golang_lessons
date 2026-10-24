package main

import "fmt"

type Profile struct {
	Name       string
	Age        int
	Email      string
	Newsletter bool
}

type ProfilePatch struct {
	Name       *string
	Age        *int
	Email      *string
	Newsletter *bool
}

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

func ptr[T any](v T) *T {
	return &v
}

func main() {
	profile := Profile{Name: "Анна", Age: 30, Email: "anna@example.com", Newsletter: true}
	fmt.Printf("%+v\n", profile)

	ApplyPatch(&profile, ProfilePatch{
		Age:        ptr(31),
		Newsletter: ptr(false),
	})
	fmt.Printf("%+v\n", profile)

	ApplyPatch(&profile, ProfilePatch{})
	fmt.Printf("%+v\n", profile)

	ApplyPatch(&profile, ProfilePatch{Age: ptr(0)})
	fmt.Printf("%+v\n", profile)

	ApplyPatch(nil, ProfilePatch{Name: ptr("Борис")})
	fmt.Println("ApplyPatch(nil, ...) отработал без паники")
}
