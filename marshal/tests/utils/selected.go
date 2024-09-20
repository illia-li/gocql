package utils

type Selected []bool

func (l *Selected) Len() int {
	return len(*l)
}

func (l *Selected) IsSelected(id int) bool {
	return (*l)[id]
}

func (l *Selected) SetSelected(id int) {
	(*l)[id] = true
}

func (l *Selected) AllSelected(ids []int) bool {
	for _, id := range ids {
		if l.IsSelected(id) {
			return true
		}
	}
	return false
}
