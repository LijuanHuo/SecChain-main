package mysockets

import (
	"strconv"
)

func GetipAddr(ip0 string, port int) string {
	pp := strconv.Itoa(port)
	var ss string = ip0 + ":" + pp
	return ss
}
