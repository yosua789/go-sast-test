package helper

import "fmt"

func ConvertRowColumnKey(row, column int) string {
	return fmt.Sprintf("R%dC%d", row, column)
}
