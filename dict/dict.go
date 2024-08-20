package dict

type Dict interface {
	Convert(word string) ([]string, error)
}
