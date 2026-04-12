package main

import (
    "fmt"

    "github.com/dipakrasal2009/DevOps-Cohort-Go/payment-interface/pkg"
    "github.com/dipakrasal2009/DevOps-Cohort-Go/payment-interface/pkg/upi"
    "github.com/dipakrasal2009/DevOps-Cohort-Go/payment-interface/pkg/creditcard"
)

func Checkout(method pkg.PaymentMethod, amount float64) string {
    return method.Pay(amount)
}

func main() {
    fmt.Println("payment interface example")

    // UPI Payment
    dipakUPI := upi.UPIPayment{
        UpiID: "dipak@okhdfc",
        App:   "Gpay",
    }

    msg1 := Checkout(dipakUPI, 50.55)
    fmt.Printf("UPI: %s\n", msg1)

    // Credit Card Payment
    dipakCard := creditcard.CreditCardPayment{
        CardNumber: "1234-5678-9012-3456",
        ExpiryDate: "12/28",
        CVV:        "123",
    }

    msg2 := Checkout(dipakCard, 1200.00)
    fmt.Printf("Credit Card: %s\n", msg2)
}
