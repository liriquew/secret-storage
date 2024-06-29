package shamir

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	mathrand "math/rand"
)

type polynomial struct {
	coefficients []uint8
}

type ShamirSecret struct {
	Parts     int `json:"parts"`
	Threshold int `json:"threshold"`
}

// Возвращает полином заданной степени с случайными коэффициентами,
// но с определенным свободным членом
func makePolynomial(intercept, degree uint8) (polynomial, error) {
	p := polynomial{
		coefficients: make([]byte, degree+1),
	}

	p.coefficients[0] = intercept

	if _, err := rand.Read(p.coefficients[1:]); err != nil {
		return p, err
	}

	return p, nil
}

// Возвращает значение полинома в точке x
// ax2 + bx + c = (ax + b)x + c
// p.coefficients = [c, b, a]
// i = 1
// out == a
// coef == b
// ----
// out == ax + b
// i = 0
// coef == c
// out == (ax + b)*x + c
func (p *polynomial) evaluate(x uint8) uint8 {
	if x == 0 {
		return p.coefficients[0]
	}

	// вычисление значения полинома по схеме горнера
	degree := len(p.coefficients) - 1
	out := p.coefficients[degree]
	for i := degree - 1; i >= 0; i-- {
		coeff := p.coefficients[i]
		out = add(mult(out, x), coeff)
	}
	return out
}

// Вычисление значения интерполяционного многочлена Лагранжа в точке x
func interpolatePolynomial(x_samples, y_samples []uint8, x uint8) uint8 {
	limit := len(x_samples)
	var result, basis uint8
	for i := 0; i < limit; i++ {
		basis = 1
		for j := 0; j < limit; j++ {
			if i == j {
				continue
			}
			num := add(x, x_samples[j])
			denom := add(x_samples[i], x_samples[j])
			term := div(num, denom)
			basis = mult(basis, term)
		}
		group := mult(y_samples[i], basis)
		result = add(result, group)
	}
	return result
}

// Деление в поле Галуа GF[256]
func div(a, b uint8) uint8 {
	ret := int(mult(a, inverse(b)))

	ret = subtle.ConstantTimeSelect(subtle.ConstantTimeByteEq(a, 0), 0, ret)
	return uint8(ret)
}

// Нахождение обратного числа в поле Галуа GF[256]
func inverse(a uint8) uint8 {
	b := mult(a, a)
	c := mult(a, b)
	b = mult(c, c)
	b = mult(b, b)
	c = mult(b, c)
	b = mult(b, b)
	b = mult(b, b)
	b = mult(b, c)
	b = mult(b, b)
	b = mult(a, b)

	return mult(b, b)
}

// Умножение чисел в поле Галуа GF[256]
func mult(a, b uint8) (out uint8) {
	var r uint8 = 0
	var i uint8 = 8

	for i > 0 {
		i--
		r = (-(b >> i & 1) & a) ^ (-(r >> 7) & 0x1B) ^ (r + r)
	}

	return r
}

// Сложение/вычитание двух чисел в поле Галуа GF[256]
func add(a, b uint8) uint8 {
	return a ^ b
}

// Генерация из секрета `частей`
// parts - число частей
// threshold - число частей, необходимых для восстановления
// 2 <= parts, recoveryParts <= 256
// Возвращаемые доли на один байт длиннее секрета, так как к ним прикрепляется метка,
// используемая для восстановления секрета.
func Split(secret []byte, parts, threshold int) ([][]byte, error) {
	if parts < threshold {
		return nil, fmt.Errorf("parts cannot be less than threshold")
	}
	if parts > 255 {
		return nil, fmt.Errorf("parts cannot exceed 255")
	}
	if threshold < 2 {
		return nil, fmt.Errorf("threshold must be at least 2")
	}
	if threshold > 255 {
		return nil, fmt.Errorf("threshold cannot exceed 255")
	}
	if len(secret) == 0 {
		return nil, fmt.Errorf("cannot split an empty secret")
	}

	// Генерация списка координат
	xCoordinates := mathrand.Perm(parts)

	// Выделение памяти под части секрета, и инициализация последних байтов смещением.
	// В общем случае каждая часть выглядит следующим образом: [y1, y2, .., yN, x]
	out := make([][]byte, parts)
	for idx := range out {
		out[idx] = make([]byte, len(secret)+1)
		out[idx][len(secret)] = uint8(xCoordinates[idx]) + 1
	}

	// Поскольку один полином может содержать только один байт секрета,
	// для каждого байта создается свой полином.
	for idx, val := range secret {
		p, err := makePolynomial(val, uint8(threshold-1))
		if err != nil {
			return nil, fmt.Errorf("failed to generate polynomial: %w", err)
		}

		// Генерация значений пар (x, y),
		for i := 0; i < parts; i++ {
			x := uint8(xCoordinates[i]) + 1
			y := p.evaluate(x)
			out[i][idx] = y
		}
	}

	return out, nil
}

// Восстановление секерта из частей
func Combine(parts [][]byte) ([]byte, error) {
	if len(parts) < 2 {
		return nil, fmt.Errorf("less than two parts cannot be used to reconstruct the secret")
	}

	// Проверка, что все части одной длины
	firstPartLen := len(parts[0])
	if firstPartLen < 2 {
		return nil, fmt.Errorf("parts must be at least two bytes")
	}
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) != firstPartLen {
			return nil, fmt.Errorf("all parts must be the same length")
		}
	}

	secret := make([]byte, firstPartLen-1)

	// буферы для хранения иходных данных для каждого байта секрета
	x_samples := make([]uint8, len(parts))
	y_samples := make([]uint8, len(parts))

	// проверка, что все значения в x_samples различные
	checkMap := map[byte]bool{}
	for i, part := range parts {
		samp := part[firstPartLen-1]
		if exists := checkMap[samp]; exists {
			return nil, fmt.Errorf("duplicate part detected")
		}
		checkMap[samp] = true
		x_samples[i] = samp
	}

	// восстановление каждого байта секрета
	for idx := range secret {
		// определение иходных данных по оси ординат
		for i, part := range parts {
			y_samples[i] = part[idx]
		}

		// вычисление значения интерполяционного многочлена в точке 0
		// так как при разбиении, каждый байт секрета использовался
		// в качестве смещения, следовательно вычисление в точке x = 0
		// даст искомое значение байта секрета
		secret[idx] = interpolatePolynomial(x_samples, y_samples, 0)
	}
	return secret, nil
}
