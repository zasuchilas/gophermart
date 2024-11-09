package main

import "github.com/zasuchilas/gophermart/internal/accrual"

func main() {
	a := accrual.New()
	a.Run()
}
