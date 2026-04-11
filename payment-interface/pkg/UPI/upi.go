package upi

import "fmt"

type UPTPayment struct {
  UpiID string
  App string
}

func (u UPIPayment) Pay(amount float64) string {
  msg := fmt.Sprintf("UPI payment if %.2f successfully complete using %s \n",amount,u.UpiID)
  return msg
}
