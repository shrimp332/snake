package util 

// dangerous, panics if err
func Unwrap[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
