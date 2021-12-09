package bsizeparser

import (
	"strconv"
	"strings"
)

func Parse(v string) (r uint64, err error) {
	v = strings.Trim(v, " \t")
	v = strings.ToLower(v)

	if len(v) == 0 {
		return
	}

	var multiplier uint64 = 1
	switch v[len(v)-1] {
	case 'k':
		multiplier = 1024
	case 'm':
		multiplier = 1024 * 1024
	case 'g':
		multiplier = 1024 * 1024 * 1024
	case 't':
		multiplier = 1024 * 1024 * 1024 * 1024
	case 'p':
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024
	}

	if multiplier != 1 {
		v = v[:len(v)-1]
	}

	if len(v) == 0 {
		r = 1
	} else {
		r, err = strconv.ParseUint(v, 10, 64)
		if err != nil {
			return
		}
	}

	r *= multiplier

	return
}
