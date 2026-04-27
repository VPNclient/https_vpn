package malva

// Таблиці для NTT (Number Theoretic Transform)
// ζ = 17 - примітивний 512-й корінь з одиниці в Z_Q

// zetas - степені ζ для NTT
var zetas = [128]int16{
	1, 1729, 2580, 3289, 2642, 630, 1897, 848,
	1062, 1919, 193, 797, 2786, 3260, 569, 1746,
	296, 2447, 1339, 1476, 3046, 56, 2240, 1333,
	1426, 2094, 535, 2882, 2393, 2879, 1974, 821,
	289, 331, 3253, 1756, 1197, 2304, 2277, 2055,
	650, 1977, 2513, 632, 2865, 33, 1320, 1915,
	2319, 1435, 807, 452, 1438, 2868, 1534, 2402,
	2647, 2617, 1481, 648, 2474, 3110, 1227, 910,
	17, 2761, 583, 2649, 1637, 723, 2288, 1100,
	1409, 2662, 3281, 233, 756, 2156, 3015, 3050,
	1703, 1651, 2789, 1789, 1847, 952, 1461, 2687,
	939, 2308, 2437, 2388, 733, 2337, 268, 641,
	1584, 2298, 2037, 3220, 375, 2549, 2090, 1645,
	1063, 319, 2773, 757, 2099, 561, 2466, 2594,
	2804, 1092, 403, 1026, 1143, 2150, 2775, 886,
	1722, 1212, 1874, 1029, 2110, 2935, 885, 2154,
}

// zetasInv - інверсні степені ζ для INTT
var zetasInv = [128]int16{
	1175, 2444, 394, 1219, 2300, 1455, 2117, 1607,
	2443, 554, 1179, 2186, 2303, 2926, 2237, 525,
	735, 863, 2768, 1230, 2572, 556, 3010, 2266,
	1684, 1239, 780, 2954, 109, 1292, 1031, 1745,
	2688, 2400, 2390, 2092, 642, 2377, 1540, 1482,
	1868, 2377, 2596, 992, 941, 892, 1021, 2390,
	642, 2596, 2377, 1868, 1482, 1540, 2377, 642,
	2092, 2390, 2400, 2688, 1745, 1031, 1292, 109,
	2954, 780, 1239, 1684, 2266, 3010, 556, 2572,
	1230, 2768, 863, 735, 525, 2237, 2926, 2303,
	2186, 1179, 554, 2443, 1607, 2117, 1455, 2300,
	1219, 394, 2444, 1175, 2154, 885, 2935, 2110,
	1029, 1874, 1212, 1722, 886, 2775, 2150, 1143,
	1026, 403, 1092, 2804, 2594, 2466, 561, 2099,
	757, 2773, 319, 1063, 1645, 2090, 2549, 375,
	3220, 2037, 2298, 1584, 641, 268, 2337, 733,
}

// NTT виконує Number Theoretic Transform над поліномом
// Перетворює з нормального представлення в NTT домен
func (p *Poly) NTT() {
	k := 1
	for l := 128; l >= 2; l >>= 1 {
		for start := 0; start < N; start += 2 * l {
			zeta := zetas[k]
			k++
			for j := start; j < start+l; j++ {
				t := montgomeryReduce(int32(zeta) * int32(p[j+l]))
				p[j+l] = p[j] - t
				p[j] = p[j] + t
			}
		}
	}
	p.reduce()
}

// InvNTT виконує інверсний NTT
// Перетворює з NTT домену назад в нормальне представлення
func (p *Poly) InvNTT() {
	k := 127
	for l := 2; l <= 128; l <<= 1 {
		for start := 0; start < N; start += 2 * l {
			zeta := zetasInv[k]
			k--
			for j := start; j < start+l; j++ {
				t := p[j]
				p[j] = t + p[j+l]
				p[j+l] = t - p[j+l]
				p[j+l] = montgomeryReduce(int32(zeta) * int32(p[j+l]))
			}
		}
	}

	// Множення на N^(-1) mod Q = 3303
	const f = 3303
	for i := 0; i < N; i++ {
		p[i] = montgomeryReduce(int32(f) * int32(p[i]))
	}
	p.reduce()
}

// reduce приводить всі коефіцієнти до діапазону [0, Q)
func (p *Poly) reduce() {
	for i := 0; i < N; i++ {
		p[i] = barrettReduce(p[i])
	}
}

// Add додає два поліноми
func (p *Poly) Add(a, b *Poly) {
	for i := 0; i < N; i++ {
		p[i] = a[i] + b[i]
	}
}

// Sub віднімає поліноми (p = a - b)
func (p *Poly) Sub(a, b *Poly) {
	for i := 0; i < N; i++ {
		p[i] = a[i] - b[i]
	}
}

// BaseMul множить два поліноми в NTT домені (базове множення)
func (p *Poly) BaseMul(a, b *Poly) {
	for i := 0; i < N/2; i++ {
		// Використовуємо zeta з таблиці (індекс mod 128)
		zeta := int32(zetas[i%128])

		// Множення пари коефіцієнтів
		a0, a1 := int32(a[2*i]), int32(a[2*i+1])
		b0, b1 := int32(b[2*i]), int32(b[2*i+1])

		p[2*i] = montgomeryReduce(a0*b0 + int32(montgomeryReduce(a1*b1))*zeta)
		p[2*i+1] = montgomeryReduce(a0*b1 + a1*b0)
	}
}

// PolyVecNTT виконує NTT над вектором поліномів
func (v *PolyVec) NTT() {
	for i := 0; i < K; i++ {
		v[i].NTT()
	}
}

// PolyVecInvNTT виконує інверсний NTT над вектором
func (v *PolyVec) InvNTT() {
	for i := 0; i < K; i++ {
		v[i].InvNTT()
	}
}

// PolyVecAdd додає два вектори поліномів
func (v *PolyVec) Add(a, b *PolyVec) {
	for i := 0; i < K; i++ {
		v[i].Add(&a[i], &b[i])
	}
}

// PolyVecReduce приводить всі коефіцієнти до діапазону [0, Q)
func (v *PolyVec) Reduce() {
	for i := 0; i < K; i++ {
		v[i].reduce()
	}
}

// InnerProduct обчислює скалярний добуток двох векторів в NTT домені
func InnerProduct(a, b *PolyVec) *Poly {
	var result Poly
	var tmp Poly

	for i := 0; i < K; i++ {
		tmp.BaseMul(&a[i], &b[i])
		result.Add(&result, &tmp)
	}

	result.reduce()
	return &result
}

// MatVecMul множить матрицю на вектор (A * v)
// Матриця A представлена як масив рядків
func MatVecMul(a *[K]PolyVec, v *PolyVec, transpose bool) *PolyVec {
	var result PolyVec

	for i := 0; i < K; i++ {
		var tmp Poly
		for j := 0; j < K; j++ {
			var prod Poly
			if transpose {
				prod.BaseMul(&a[j][i], &v[j])
			} else {
				prod.BaseMul(&a[i][j], &v[j])
			}
			tmp.Add(&tmp, &prod)
		}
		result[i] = tmp
	}

	result.Reduce()
	return &result
}
