package main

import "fmt"

type CustomerStatus int

const (
	Regular CustomerStatus = iota
	Silver
	Gold
)

const (
	DeliveryCost    = 299
	FreeDeliveryMin = 3000
)

func getStatusName(status CustomerStatus) string {
	switch status {
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

func getDiscountPercentage(status CustomerStatus) int {
	switch status {
	case Regular:
		return 0
	case Silver:
		return 5
	case Gold:
		return 10
	default:
		fmt.Println("Unknown")
		return 0
	}
}

func calculateDelivery(amountAfterDiscount int) int {
	if amountAfterDiscount >= FreeDeliveryMin {
		return 0
	}
	return DeliveryCost
}

func main() {

}
