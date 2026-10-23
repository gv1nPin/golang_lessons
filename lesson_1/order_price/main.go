package main

import "fmt"

type CustomerStatus int

const (
	Regular CustomerStatus = iota
	Silver
	Gold
)

const (
	deliveryCost    = 299.0
	freeDeliveryMin = 3000.0
)

func statusName(status CustomerStatus) string {
	switch status {
	case Regular:
		return "Regular"
	case Silver:
		return "Silver"
	case Gold:
		return "Gold"
	default:
		return "неизвестный статус"
	}
}

func discountPercent(status CustomerStatus) float64 {
	switch status {
	case Regular:
		return 0
	case Silver:
		return 5
	case Gold:
		return 10
	default:
		return 0
	}
}

func calculateDelivery(amountAfterDiscount float64) float64 {
	if amountAfterDiscount >= freeDeliveryMin {
		return 0
	}
	return deliveryCost
}

func processOrder(amount float64, status CustomerStatus) {
	discount := amount * discountPercent(status) / 100
	amountAfterDiscount := amount - discount
	delivery := calculateDelivery(amountAfterDiscount)
	total := amountAfterDiscount + delivery

	fmt.Printf("Сумма: %.2f, статус: %s\n", amount, statusName(status))
	fmt.Printf("Скидка: %.2f\n", discount)
	fmt.Printf("Доставка: %.2f\n", delivery)
	fmt.Printf("Итого: %.2f\n\n", total)
}

func main() {
	processOrder(4500, Gold)
	processOrder(2900, Silver)
	processOrder(1000, Regular)
	processOrder(1000, CustomerStatus(42)) // неизвестный статус
}
