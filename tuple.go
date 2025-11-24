package soutils

type Tuple1[T1 any] struct {
	Item1 T1
}

func MakeT1[T1 any](v1 T1) Tuple1[T1] {
	return Tuple1[T1]{v1}
}

func (t *Tuple1[T1]) Values() []any {
	return []any{t.Item1}
}

func (t *Tuple1[T1]) Unpack() T1 {
	return t.Item1
}

type Tuple2[T1, T2 any] struct {
	Item1 T1
	Item2 T2
}

func MakeT2[T1, T2 any](v1 T1, v2 T2) Tuple2[T1, T2] {
	return Tuple2[T1, T2]{v1, v2}
}

func (t *Tuple2[T1, T2]) Values() []any {
	return []any{t.Item1, t.Item2}
}

func (t *Tuple2[T1, T2]) Unpack() (T1, T2) {
	return t.Item1, t.Item2
}

type Tuple3[T1, T2, T3 any] struct {
	Item1 T1
	Item2 T2
	Item3 T3
}

func MakeT3[T1, T2, T3 any](v1 T1, v2 T2, v3 T3) Tuple3[T1, T2, T3] {
	return Tuple3[T1, T2, T3]{v1, v2, v3}
}

func (t *Tuple3[T1, T2, T3]) Values() []any {
	return []any{t.Item1, t.Item2, t.Item3}
}

func (t *Tuple3[T1, T2, T3]) Unpack() (T1, T2, T3) {
	return t.Item1, t.Item2, t.Item3
}

type Tuple4[T1, T2, T3, T4 any] struct {
	Item1 T1
	Item2 T2
	Item3 T3
	Item4 T4
}

func MakeT4[T1, T2, T3, T4 any](v1 T1, v2 T2, v3 T3, v4 T4) Tuple4[T1, T2, T3, T4] {
	return Tuple4[T1, T2, T3, T4]{v1, v2, v3, v4}
}

func (t *Tuple4[T1, T2, T3, T4]) Values() []any {
	return []any{t.Item1, t.Item2, t.Item3, t.Item4}
}

func (t *Tuple4[T1, T2, T3, T4]) Unpack() (T1, T2, T3, T4) {
	return t.Item1, t.Item2, t.Item3, t.Item4
}

type Tuple5[T1, T2, T3, T4, T5 any] struct {
	Item1 T1
	Item2 T2
	Item3 T3
	Item4 T4
	Item5 T5
}

func MakeT5[T1, T2, T3, T4, T5 any](
	v1 T1, v2 T2, v3 T3, v4 T4, v5 T5) Tuple5[T1, T2, T3, T4, T5] {
	return Tuple5[T1, T2, T3, T4, T5]{v1, v2, v3, v4, v5}
}

func (t *Tuple5[T1, T2, T3, T4, T5]) Values() []any {
	return []any{t.Item1, t.Item2, t.Item3, t.Item4, t.Item5}
}

func (t *Tuple5[T1, T2, T3, T4, T5]) Unpack() (T1, T2, T3, T4, T5) {
	return t.Item1, t.Item2, t.Item3, t.Item4, t.Item5
}

type Tuple6[T1, T2, T3, T4, T5, T6 any] struct {
	Item1 T1
	Item2 T2
	Item3 T3
	Item4 T4
	Item5 T5
	Item6 T6
}

func MakeT6[T1, T2, T3, T4, T5, T6 any](
	v1 T1, v2 T2, v3 T3, v4 T4, v5 T5, v6 T6) Tuple6[T1, T2, T3, T4, T5, T6] {
	return Tuple6[T1, T2, T3, T4, T5, T6]{v1, v2, v3, v4, v5, v6}
}

func (t *Tuple6[T1, T2, T3, T4, T5, T6]) Values() []any {
	return []any{t.Item1, t.Item2, t.Item3, t.Item4, t.Item5, t.Item6}
}

func (t *Tuple6[T1, T2, T3, T4, T5, T6]) Unpack() (T1, T2, T3, T4, T5, T6) {
	return t.Item1, t.Item2, t.Item3, t.Item4, t.Item5, t.Item6
}

type Tuple7[T1, T2, T3, T4, T5, T6, T7 any] struct {
	Item1 T1
	Item2 T2
	Item3 T3
	Item4 T4
	Item5 T5
	Item6 T6
	Item7 T7
}

func MakeT7[T1, T2, T3, T4, T5, T6, T7 any](
	v1 T1, v2 T2, v3 T3, v4 T4, v5 T5, v6 T6, v7 T7) Tuple7[T1, T2, T3, T4, T5, T6, T7] {
	return Tuple7[T1, T2, T3, T4, T5, T6, T7]{v1, v2, v3, v4, v5, v6, v7}
}

func (t *Tuple7[T1, T2, T3, T4, T5, T6, T7]) Values() []any {
	return []any{t.Item1, t.Item2, t.Item3, t.Item4, t.Item5, t.Item6, t.Item7}
}

func (t *Tuple7[T1, T2, T3, T4, T5, T6, T7]) Unpack() (T1, T2, T3, T4, T5, T6, T7) {
	return t.Item1, t.Item2, t.Item3, t.Item4, t.Item5, t.Item6, t.Item7
}
