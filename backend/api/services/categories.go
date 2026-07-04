package services

import "slices"

var movieCategoryIDs = []int{8, 9, 11, 37, 43, 14, 12, 13, 47, 15, 29, 34, 36}
var tvCategoryIDs = []int{26, 32, 27, 44}

func ResolveQbtCategory(categoryID int) (string, bool) {
	if slices.Contains(movieCategoryIDs, categoryID) {
		return "movies", true
	}

	if slices.Contains(tvCategoryIDs, categoryID) {
		return "tvseries", true
	}

	return "", false
}

func IsMovieOrTVCategory(categoryID int) bool {
	_, ok := ResolveQbtCategory(categoryID)
	return ok
}
