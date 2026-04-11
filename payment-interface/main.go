package main

import (
        "fmt"
        "github.com/dipakrasal2009/DevOps-Cohort-GO/payment-interface/pkg/upi"

func Checkout(method payments,PaymentMethod, amount float64) string {
string  msg := method.pay(amount)
  return msg

}

func main() {
  fmt.Println("payment interface example")

  dipakUPI := upi.UPIPayment{UpiID: "dipak@okhdfc",App: Gpay}
  msg = Checkout(dipakUPI, 50.55)
  fmt.Println("payment successful : %s", msg)
}
