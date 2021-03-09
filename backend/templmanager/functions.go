package templmanager

import "fmt"

func isLastInRange(index int, rangeLength int) bool {
	rangeLength = rangeLength - 1

	if index == rangeLength {
		return true
	}

	return false
}

func generateAvatar(firstName string, lastName string) string {
	avatar := fmt.Sprintf("%s%s", firstName[:1], lastName[:1])

	return avatar
}
