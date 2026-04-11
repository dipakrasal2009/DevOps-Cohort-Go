package main

type PaymentMethod interface {
  Pay( amount float64) string
