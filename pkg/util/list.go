package util

import (
	"container/list"
)

// Get return element at index of list
func Get(l *list.List, index int) *list.Element {
	if nil == l || l.Len() == 0 {
		return nil
	}

	i := 0
	for iter := l.Front(); iter != nil; iter = iter.Next() {
		if i == index {
			return iter
		}

		i++
	}

	return nil
}

// IndexOf return index of element in list
func IndexOf(l *list.List, value interface{}) int {
	i := 0
	for iter := l.Front(); iter != nil; iter = iter.Next() {
		if iter.Value == value {
			return i
		}

		i++
	}

	return -1
}

// Remove remove from list
func Remove(l *list.List, value interface{}) {
	var e *list.Element

	for iter := l.Front(); iter != nil; iter = iter.Next() {
		if iter.Value == value {
			e = iter
			break
		}
	}

	if nil != e {
		l.Remove(e)
	}
}

// ToStringArray return string array
func ToStringArray(l *list.List) []string {
	if nil == l {
		return nil
	}

	values := make([]string, l.Len())

	i := 0
	for iter := l.Front(); iter != nil; iter = iter.Next() {
		s, _ := iter.Value.(string)
		values[i] = s

		i++
	}

	return values
}
