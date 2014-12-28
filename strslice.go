package uniconfig

import (
	"fmt"
	"reflect"
	"strings"
)

type strslice struct {
	data *[]string
}

func NewStrSlice(data *[]string) *strslice {
	return &strslice{data}
}

func (i *strslice) String() string {
	return fmt.Sprintf("%+v", *i.data)
}

// The second method is Set(value string) error
func (i *strslice) Set(value string) error {
	tmp := []string{}
	res := strings.Split(value, ",")
	for _, p := range res {
		val := strings.TrimSpace(p)
		tmp = append(tmp, val)
	}
	*i.data = tmp

	return nil
}

var strSliceType = reflect.TypeOf([]string{})
