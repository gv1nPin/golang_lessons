package main

import "fmt"

type CustomerStatus int

const (
	Regular CustomerStatus = iota
	Silver
	Gold
)

const (
	deliveryCost    = 299
	freeDeliveryMin = 3000
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

func calculateDelivery(amountAfterDiscount float64) float64 {
	if amountAfterDiscount >= freeDeliveryMin {
		return 0
	}
	return deliveryCost
}

func main() {
	var initialAmount float64
	var status CustomerStatus
	var discountPercentage int
	var discountAmount float64
	var amountAfterDiscount float64
	var deliveryCost float64
	var totalAmount float64

	initialAmount = 4500
	status = Gold
	discountPercentage = getDiscountPercentage(status)
	discountAmount = initialAmount * (float64(discountPercentage) / 100)
	amountAfterDiscount = initialAmount - discountAmount
	deliveryCost = calculateDelivery(amountAfterDiscount)
	totalAmount = initialAmount - discountAmount

	fmt.Printf("Сумма: %.2f, статус: %s\n", initialAmount, getStatusName(status))
	fmt.Printf("Скидка: %.2f\n", discountAmount)
	fmt.Printf("Доставка: %.2f\n", deliveryCost)
	fmt.Printf("Итого: %.2f\n\n", totalAmount)

}
