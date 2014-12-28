package main

import (
	"github.com/Babazka/uniconfig"
	"log"
)

type MyConfig struct {
	Debug     bool
	Count     int `help:"number of items"`
	Intlist   []int
	Strlist   []string
	Floatlist []float64
	Nested1   struct {
		A       string
		B       string
		ignored string
	}
	Nested2 struct {
		Zzz bool
	}
}

func main() {
	config := &MyConfig{Intlist: []int{1, 2, 3}}
	config.Count = 42
	uniconfig.Load(config)
	log.Printf("Hi")
	log.Printf("Final config: \n%s", uniconfig.ConfigAsIniFile(config))
}
