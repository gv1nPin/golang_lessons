package main

import (
	"clinic/delivery"
	"clinic/notifier"
	"clinic/repository"
	"clinic/usecase"
)

var (
	_ usecase.Repository = (*repository.Memory)(nil)
	_ usecase.Repository = (*repository.Postgres)(nil)
	_ usecase.Notifier   = (*notifier.SMS)(nil)
	_ delivery.Usecase   = (*usecase.Booking)(nil)
)
