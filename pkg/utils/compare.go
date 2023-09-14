package utils

func DiffArrays(a, b []string) []string {
	var diff []string
	for _, v := range a {
		found := false
		for _, w := range b {
			if v == w {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, v)
		}
	}
	return diff
}

func CommonArray(a, b []string) []string {
	var common []string
	for _, v := range a {
		for _, w := range b {
			if v == w {
				common = append(common, v)
			}
		}
	}
	return common
}
