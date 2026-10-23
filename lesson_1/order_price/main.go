package main

import "fmt"

type customerStatus int

const (
	Regular customerStatus = iota
	Silver
	Gold
)

const (
	deliveryCost    = 299
	freeDeliveryMin = 3000
)

func getStatusName(status customerStatus) string {
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

func getDiscountPercentage(status customerStatus) int {
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

func processOrder(initialAmount float64, status customerStatus) {
	discountPercentage := getDiscountPercentage(status)
	discountAmount := initialAmount * (float64(discountPercentage) / 100)
	amountAfterDiscount := initialAmount - discountAmount
	currentDeliveryCost := calculateDelivery(amountAfterDiscount)
	totalAmount := amountAfterDiscount + currentDeliveryCost

	fmt.Printf("Сумма: %.2f, статус: %s\n", initialAmount, getStatusName(status))
	fmt.Printf("Скидка: %.2f\n", discountAmount)
	fmt.Printf("Доставка: %.2f\n", currentDeliveryCost)
	fmt.Printf("Итого: %.2f\n\n", totalAmount)
}

func main() {
	processOrder(4500, Gold)
	processOrder(2900, Silver)
	processOrder(1000, Regular)
}
