package interpreter

type queue struct {
	v []foreachBuffer
	cursor
}

func (q *queue) push(v foreachBuffer) {
	q.v = append(q.v, v)
}

func (q *queue) mv(v int) {
	vLen := len(q.v)
	if vLen == 0 || q.p+v >= vLen {
		return
	}
	q.mvTo(q.p + v)
}

func (q *queue) mvTo(v int) {
	q.p = v
}

func (q *queue) last() *foreachBuffer {
	return q.lastN(0)
}

func (q *queue) firstN(n int) *foreachBuffer {
	return &q.v[n]
}

func (q *queue) lastN(n int) *foreachBuffer {
	return &q.v[len(q.v)-1-n]
}

func (q *queue) load() *foreachBuffer {
	return q.loadN(q.p)
}

func (q *queue) loadN(n int) *foreachBuffer {
	return &q.v[n]
}

func (q *queue) len() int {
	return len(q.v)
}
