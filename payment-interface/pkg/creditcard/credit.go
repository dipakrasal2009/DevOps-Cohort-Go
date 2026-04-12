package creditcard

import "fmt"

type CreditCardPayment struct {
    CardNumber string
    ExpiryDate string
    CVV        string
}

func (c CreditCardPayment) Pay(amount float64) string {
    return fmt.Sprintf(
        "Credit card payment of %.2f successful using %s",
        amount,
        c.CardNumber,
    )
}
