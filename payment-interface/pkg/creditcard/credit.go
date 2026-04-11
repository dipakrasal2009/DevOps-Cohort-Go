package creditcard

import "fmt"

type CredirCardPayment struct {
  CardNumber string
  ExpiryDate string
  CVV string
}

func (c CreditCardPayment) Pay(amount float64) string{
  msg := fmt.Sprintf("credit card pament of %.2f successfully completed uisng %s \n",amount,c.CardNumber)
  return msg
}
