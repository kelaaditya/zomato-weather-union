package utilities

// generic function to check if a pointer is nil.
// if the pointer is not nil, dereference it to get value.
func DereferenceOrNil[T comparable](p *T) any {
	if p != nil {
		return *p
	}
	return nil
}
