// Package envx provides an ordered environment-variable collection.
package envx

// Ordered collects environment variables in first-set order, with the last
// value assigned to a key winning. This mirrors how a process environment is
// built: base entries first, then overrides, where later values replace
// earlier ones while keeping the key's original position.
type Ordered struct {
	values map[string]string
	order  []string
}

// New returns an empty Ordered.
func New() *Ordered {
	return &Ordered{values: make(map[string]string)}
}

// Set assigns value to key, appending key to the order the first time it is
// seen.
func (o *Ordered) Set(key, value string) {
	if _, exists := o.values[key]; !exists {
		o.order = append(o.order, key)
	}
	o.values[key] = value
}

// Len returns the number of distinct keys.
func (o *Ordered) Len() int {
	return len(o.order)
}

// Slice returns the environment as "KEY=VALUE" strings in first-set order.
func (o *Ordered) Slice() []string {
	out := make([]string, 0, len(o.order))
	for _, k := range o.order {
		out = append(out, k+"="+o.values[k])
	}
	return out
}
