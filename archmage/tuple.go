package archmage

// Tuple1 holds a single typed value.
type Tuple1[T1 any] struct {
	Item1 T1 `json:"item1"`
}

// MakeTuple1 creates a Tuple1 from a single value.
func MakeTuple1[T1 any](v1 T1) Tuple1[T1] {
	return Tuple1[T1]{v1}
}

// Values returns all tuple items as a slice.
func (t *Tuple1[T1]) Values() []any {
	return []any{t.Item1}
}

// Unpack returns the tuple's value.
func (t *Tuple1[T1]) Unpack() T1 {
	return t.Item1
}

// Tuple2 holds two typed values.
type Tuple2[T1, T2 any] struct {
	Item1 T1 `json:"item1"`
	Item2 T2 `json:"item2"`
}

// MakeTuple2 creates a Tuple2 from two values.
func MakeTuple2[T1, T2 any](v1 T1, v2 T2) Tuple2[T1, T2] {
	return Tuple2[T1, T2]{v1, v2}
}

// Values returns all tuple items as a slice.
func (t *Tuple2[T1, T2]) Values() []any {
	return []any{t.Item1, t.Item2}
}

// Unpack returns the tuple's values.
func (t *Tuple2[T1, T2]) Unpack() (T1, T2) {
	return t.Item1, t.Item2
}

// Tuple3 holds three typed values.
type Tuple3[T1, T2, T3 any] struct {
	Item1 T1 `json:"item1"`
	Item2 T2 `json:"item2"`
	Item3 T3 `json:"item3"`
}

// MakeTuple3 creates a Tuple3 from three values.
func MakeTuple3[T1, T2, T3 any](v1 T1, v2 T2, v3 T3) Tuple3[T1, T2, T3] {
	return Tuple3[T1, T2, T3]{v1, v2, v3}
}

// Values returns all tuple items as a slice.
func (t *Tuple3[T1, T2, T3]) Values() []any {
	return []any{t.Item1, t.Item2, t.Item3}
}

// Unpack returns the tuple's values.
func (t *Tuple3[T1, T2, T3]) Unpack() (T1, T2, T3) {
	return t.Item1, t.Item2, t.Item3
}

// Tuple4 holds four typed values.
type Tuple4[T1, T2, T3, T4 any] struct {
	Item1 T1 `json:"item1"`
	Item2 T2 `json:"item2"`
	Item3 T3 `json:"item3"`
	Item4 T4 `json:"item4"`
}

// MakeTuple4 creates a Tuple4 from four values.
func MakeTuple4[T1, T2, T3, T4 any](v1 T1, v2 T2, v3 T3, v4 T4) Tuple4[T1, T2, T3, T4] {
	return Tuple4[T1, T2, T3, T4]{v1, v2, v3, v4}
}

// Values returns all tuple items as a slice.
func (t *Tuple4[T1, T2, T3, T4]) Values() []any {
	return []any{t.Item1, t.Item2, t.Item3, t.Item4}
}

// Unpack returns the tuple's values.
func (t *Tuple4[T1, T2, T3, T4]) Unpack() (T1, T2, T3, T4) {
	return t.Item1, t.Item2, t.Item3, t.Item4
}

// Tuple5 holds five typed values.
type Tuple5[T1, T2, T3, T4, T5 any] struct {
	Item1 T1 `json:"item1"`
	Item2 T2 `json:"item2"`
	Item3 T3 `json:"item3"`
	Item4 T4 `json:"item4"`
	Item5 T5 `json:"item5"`
}

// MakeTuple5 creates a Tuple5 from five values.
func MakeTuple5[T1, T2, T3, T4, T5 any](
	v1 T1, v2 T2, v3 T3, v4 T4, v5 T5) Tuple5[T1, T2, T3, T4, T5] {
	return Tuple5[T1, T2, T3, T4, T5]{v1, v2, v3, v4, v5}
}

// Values returns all tuple items as a slice.
func (t *Tuple5[T1, T2, T3, T4, T5]) Values() []any {
	return []any{t.Item1, t.Item2, t.Item3, t.Item4, t.Item5}
}

// Unpack returns the tuple's values.
func (t *Tuple5[T1, T2, T3, T4, T5]) Unpack() (T1, T2, T3, T4, T5) {
	return t.Item1, t.Item2, t.Item3, t.Item4, t.Item5
}

// Tuple6 holds six typed values.
type Tuple6[T1, T2, T3, T4, T5, T6 any] struct {
	Item1 T1 `json:"item1"`
	Item2 T2 `json:"item2"`
	Item3 T3 `json:"item3"`
	Item4 T4 `json:"item4"`
	Item5 T5 `json:"item5"`
	Item6 T6 `json:"item6"`
}

// MakeTuple6 creates a Tuple6 from six values.
func MakeTuple6[T1, T2, T3, T4, T5, T6 any](
	v1 T1, v2 T2, v3 T3, v4 T4, v5 T5, v6 T6) Tuple6[T1, T2, T3, T4, T5, T6] {
	return Tuple6[T1, T2, T3, T4, T5, T6]{v1, v2, v3, v4, v5, v6}
}

// Values returns all tuple items as a slice.
func (t *Tuple6[T1, T2, T3, T4, T5, T6]) Values() []any {
	return []any{t.Item1, t.Item2, t.Item3, t.Item4, t.Item5, t.Item6}
}

// Unpack returns the tuple's values.
func (t *Tuple6[T1, T2, T3, T4, T5, T6]) Unpack() (T1, T2, T3, T4, T5, T6) {
	return t.Item1, t.Item2, t.Item3, t.Item4, t.Item5, t.Item6
}

// Tuple7 holds seven typed values.
type Tuple7[T1, T2, T3, T4, T5, T6, T7 any] struct {
	Item1 T1 `json:"item1"`
	Item2 T2 `json:"item2"`
	Item3 T3 `json:"item3"`
	Item4 T4 `json:"item4"`
	Item5 T5 `json:"item5"`
	Item6 T6 `json:"item6"`
	Item7 T7 `json:"item7"`
}

// MakeTuple7 creates a Tuple7 from seven values.
func MakeTuple7[T1, T2, T3, T4, T5, T6, T7 any](
	v1 T1, v2 T2, v3 T3, v4 T4, v5 T5, v6 T6, v7 T7) Tuple7[T1, T2, T3, T4, T5, T6, T7] {
	return Tuple7[T1, T2, T3, T4, T5, T6, T7]{v1, v2, v3, v4, v5, v6, v7}
}

// Values returns all tuple items as a slice.
func (t *Tuple7[T1, T2, T3, T4, T5, T6, T7]) Values() []any {
	return []any{t.Item1, t.Item2, t.Item3, t.Item4, t.Item5, t.Item6, t.Item7}
}

// Unpack returns the tuple's values.
func (t *Tuple7[T1, T2, T3, T4, T5, T6, T7]) Unpack() (T1, T2, T3, T4, T5, T6, T7) {
	return t.Item1, t.Item2, t.Item3, t.Item4, t.Item5, t.Item6, t.Item7
}
