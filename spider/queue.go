package spider

type StringQueue struct {
	q []string
}

func (sq *StringQueue) Insert(str string) {
	sq.q = append(sq.q, str)
}

func (sq *StringQueue) Pop() string {
	str := sq.q[0]
	sq.q = sq.q[1:]
	return str
}

func (sq StringQueue) Len() int {
	return len(sq.q)
}