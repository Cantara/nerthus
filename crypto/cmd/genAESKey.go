package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func main() {
	c := 32
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(base64.StdEncoding.EncodeToString((b)))
	fmt.Println(b)
	bb, _ := base64.StdEncoding.DecodeString(base64.StdEncoding.EncodeToString((b)))
	fmt.Println(bb)
	fmt.Println(len(bb))
}
