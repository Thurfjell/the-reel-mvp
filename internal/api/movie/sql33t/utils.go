package sql33t

import (
	"strconv"
	"strings"
)

func csvToInts(csv string) (ints []int) {
	if csv == "" {
		return
	}

	parts := strings.Split(csv, ",")

	ints = make([]int, 0, len(parts))

	for _, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			continue
		}
		ints = append(ints, n)
	}

	return
}

func intsToCsv(ints []int) (csv string) {
	var sb strings.Builder
	for i, n := range ints {
		if i == (len(ints) - 1) {
			sb.WriteString(strconv.Itoa(n))
			continue
		}
		sb.WriteString(strconv.Itoa(n))
		sb.WriteString(",")
	}

	csv = sb.String()
	return
}
