package main

import (
	"log"
	"github.com/Babazka/uniconfig"
)

type MyConfig struct {
	Debug   bool
	Count   int `help:"number of items"`
	Nested1 struct {
		A       string
		B       string
		ignored string
	}
	Nested2 struct {
		Zzz bool
	}
}

func main() {
	config := &MyConfig{}
	config.Count = 42
	uniconfig.Load(config)
	log.Printf("Hi")
	log.Printf("Final config: \n%s", uniconfig.ConfigAsIniFile(config))
}
