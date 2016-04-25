package util

import (
	"container/list"
)

func Get(l *list.List, index int) *list.Element {
	if nil == l || l.Len() == 0 {
		return nil
	}

	i := 0
	for iter := l.Front(); iter != nil; iter = iter.Next() {
		if i == index {
			return iter
		}

		i += 1
	}

	return nil
}

func IndexOf(l *list.List, value interface{}) int {
	i := 0
	for iter := l.Front(); iter != nil; iter = iter.Next() {
		if iter.Value == value {
			return i
		}

		i += 1
	}

	return -1
}

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

func ToStringArray(l *list.List) []string {
	if nil == l {
		return nil
	}

	values := make([]string, l.Len())

	i := 0
	for iter := l.Front(); iter != nil; iter = iter.Next() {
		s, _ := iter.Value.(string)
		values[i] = s

		i += 1
	}

	return values
}
