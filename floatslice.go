package uniconfig

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type floatslice struct {
	data *[]float64
}

func NewFloatSlice(data *[]float64) *floatslice {
	return &floatslice{data}
}

func (i *floatslice) String() string {
	return fmt.Sprintf("%d", *i.data)
}

// The second method is Set(value string) error
func (i *floatslice) Set(value string) error {
	tmp := []float64{}
	res := strings.Split(value, ",")
	for _, p := range res {
		p1 := strings.TrimSpace(p)
		val, err := strconv.ParseFloat(p1, 64)
		if err != nil {
			return fmt.Errorf("Cannot parse %s as []float64\n", value)
		}

		tmp = append(tmp, val)
	}
	*i.data = tmp

	return nil
}

var floatSliceType = reflect.TypeOf([]float64{})
