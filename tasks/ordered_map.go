package tasks

type OrderedStringTaskMap struct {
	data map[string]Task
	keys []string
}

func (o *OrderedStringTaskMap) Get(k string) Task {
	if len(o.data) == 0 {
		return nil
	}

	return o.data[k]
}

func (o *OrderedStringTaskMap) Set(k string, v Task) {
	if len(o.data) == 0 {
		o.data = make(map[string]Task)
	}

	o.data[k] = v
	o.keys = append(o.keys, k)
}

func (o *OrderedStringTaskMap) Keys() []string {
	return o.keys
}
