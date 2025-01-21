package placeholder

func Add(nums ...int) int {
	res := 0
	for _, i := range nums {
		res += i
	}
	return res
}
