package sendNodeList

func getMapKeys[T comparable, U any](m map[T]U) []T {
	keys := make([]T, len(m))
	i := 0
	for e := range m {
		keys[i] = e
		i++
	}
	return keys
}

func getSetElements[T comparable](m map[T]struct{}) []T {
	return getMapKeys(m)
}

func toSet[T comparable](xs []T) map[T]struct{} {
	set := make(map[T]struct{})
	for _, x := range xs {
		set[x] = struct{}{}
	}
	return set
}

func mergeInSet[T comparable](xs []T, ys []T) map[T]struct{} {
	set := map[T]struct{}{}
	for _, x := range xs {
		set[x] = struct{}{}
	}
	for _, y := range ys {
		set[y] = struct{}{}
	}
	return set
}
