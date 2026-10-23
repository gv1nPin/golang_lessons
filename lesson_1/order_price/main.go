package main

import "go/ast"

type CustomerStatus int

const (
	Regular CustomerStatus = iota
	Silver
	Gold
)

func main() {

	func (s CustomerStatus) String() string{
		switch s {
		case Regular:
			return "Regular"
		case Silver:
			return "Silver"
		case Gold:
			return "Gold"
		default:
			return "Unknown"
	}

	}
}
