package uniconfig

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type intslice struct {
	data *[]int
}

func NewIntSlice(data *[]int) *intslice {
	return &intslice{data}
}

func (i *intslice) String() string {
	return fmt.Sprintf("%d", *i.data)
}

// The second method is Set(value string) error
func (i *intslice) Set(value string) error {
	tmp := []int{}
	res := strings.Split(value, ",")
	for _, p := range res {
		p1 := strings.TrimSpace(p)
		val, err := strconv.Atoi(p1)
		if err != nil {
			return fmt.Errorf("Cannot parse %s as []int\n", value)
		}

		tmp = append(tmp, val)
	}
	*i.data = tmp

	return nil
}

var intSliceType = reflect.TypeOf([]int{})
