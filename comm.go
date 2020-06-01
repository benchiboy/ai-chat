package main

import (
	"log"
)

func PrintHead(a ...interface{}) {
	log.Println("========》", a)
}

func PrintTail(a ...interface{}) {
	log.Println("《========", a)
}

func PrintInfo(a ...interface{}) {
	log.Println("--->", a)
}
