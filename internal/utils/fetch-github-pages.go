package utils

import "github.com/google/go-github/v63/github"

// Parsing a Github response and going threw everything to return all
func FetchPages[T any](f func(int) ([]T, *github.Response, error)) ([]T, error) {
	d := []T{}
	pageNumber := 1
	for {
		data, res, err := f(pageNumber)

		if err != nil {
			return d, err
		}

		d = append(
			d,
			data...,
		)

		if res.NextPage == 0 {
			break
		} else {
			pageNumber += 1
		}
	}

	return d, nil
}
