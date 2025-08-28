package tasks

// OrderedStringTaskMap is a map of tasks keyed by string
type OrderedStringTaskMap struct {
	// data is the map of tasks
	data map[string]Task

	// keys is the list of keys
	keys []string
}

// Get gets the task by key
func (o *OrderedStringTaskMap) Get(k string) Task {
	if len(o.data) == 0 {
		return nil
	}

	return o.data[k]
}

// Set sets the task by key
func (o *OrderedStringTaskMap) Set(k string, v Task) {
	if len(o.data) == 0 {
		o.data = make(map[string]Task)
	}

	o.data[k] = v
	o.keys = append(o.keys, k)
}

// Keys returns the list of keys
func (o *OrderedStringTaskMap) Keys() []string {
	return o.keys
}
