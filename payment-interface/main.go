package main

import (
    "fmt"

    "github.com/dipakrasal2009/DevOps-Cohort-Go/payment-interface/pkg"
    "github.com/dipakrasal2009/DevOps-Cohort-Go/payment-interface/pkg/upi"
)

func Checkout(method pkg.PaymentMethod, amount float64) string {
    return method.Pay(amount)
}

func main() {
    fmt.Println("payment interface example")

    dipakUPI := upi.UPIPayment{
        UpiID: "dipak@okhdfc",
        App:   "Gpay",
    }

    msg := Checkout(dipakUPI, 50000.55)
    fmt.Printf("payment successful: %s\n", msg)

    yashrajUPI := upi.UPIPayment{
      UpiID: "yash@okhdfc",
      App:   "Gpay",
    }
    msg1 := Checkout(yashrajUPI, 50)
    fmt.Printf("payment successful: %s\n", msg1)

}
