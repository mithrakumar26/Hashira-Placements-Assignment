package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"
)

type RootEntry struct {
	Base  string `json:"base"`
	Value string `json:"value"`
}

type Keys struct {
	N int `json:"n"`
	K int `json:"k"`
}

func main() {
	filename := "input.json"
	if len(os.Args) >= 2 {
		filename = os.Args[1]
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading file %s: %v\n", filename, err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Fatalf("Error parsing JSON: %v\n", err)
	}

	var keysObj struct {
		Keys Keys `json:"keys"`
	}
	if err := json.Unmarshal(data, &keysObj); err != nil {
		log.Fatalf("Error parsing keys: %v\n", err)
	}
	n := keysObj.Keys.N
	k := keysObj.Keys.K

	rootMap := make(map[string]RootEntry)
	for key, rawv := range raw {
		if key == "keys" {
			continue
		}
		var e RootEntry
		if err := json.Unmarshal(rawv, &e); err != nil {
			rootMap[key] = RootEntry{}
			continue
		}
		rootMap[key] = e
	}

	var problems []string
	if len(rootMap) != n {
		problems = append(problems, fmt.Sprintf("Number of root entries found (%d) does not match 'n' (%d).", len(rootMap), n))
	}

	if k > n {
		problems = append(problems, fmt.Sprintf("'k' (%d) is greater than 'n' (%d).", k, n))
	}

	convertedRoots := make([]*big.Int, 0, len(rootMap))
	keys := make([]string, 0, len(rootMap))
	for key := range rootMap {
		keys = append(keys, key)
	}

	// numeric bubble sort starts here
	parseDecimalToInt := func(s string) (int, error) {
		if s == "" {
			return 0, fmt.Errorf("empty string")
		}
		sign := 1
		i := 0
		if s[0] == '+' {
			i = 1
		} else if s[0] == '-' {
			sign = -1
			i = 1
		}
		res := 0
		for ; i < len(s); i++ {
			ch := s[i]
			if ch < '0' || ch > '9' {
				return 0, fmt.Errorf("invalid decimal character '%c' in %q", ch, s)
			}
			d := int(ch - '0')
			res = res*10 + d
		}
		return sign * res, nil
	}

	keyInts := make([]int, len(keys))
	for i, ks := range keys {
		v, err := parseDecimalToInt(ks)
		if err != nil {
			keyInts[i] = 0
		} else {
			keyInts[i] = v
		}
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keyInts[j] < keyInts[i] {
				keyInts[i], keyInts[j] = keyInts[j], keyInts[i]
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	// numeric bubble sort ends here

	convertValueToBig := func(value string, base int) (*big.Int, error) {
		v := strings.TrimSpace(value)
		if v == "" {
			return nil, fmt.Errorf("empty value")
		}
		sign := 1
		if v[0] == '+' {
			v = v[1:]
		} else if v[0] == '-' {
			sign = -1
			v = v[1:]
		}
		if v == "" {
			return nil, fmt.Errorf("empty after sign")
		}

		res := big.NewInt(0)
		bigBase := big.NewInt(int64(base))
		for i := 0; i < len(v); i++ {
			ch := v[i]
			var digit int
			switch {
			case ch >= '0' && ch <= '9':
				digit = int(ch - '0')
			case ch >= 'A' && ch <= 'F':
				digit = int(10 + (ch - 'A'))
			case ch >= 'a' && ch <= 'f':
				digit = int(10 + (ch - 'a'))
			default:
				return nil, fmt.Errorf("invalid character '%c' in value %q", ch, value)
			}
			if digit >= base {
				return nil, fmt.Errorf("digit %d (char '%c') >= base %d in value %q", digit, ch, base, value)
			}
			res.Mul(res, bigBase)
			res.Add(res, big.NewInt(int64(digit)))
		}
		if sign < 0 {
			res.Neg(res)
		}
		return res, nil
	}
	for _, key := range keys {
		entry := rootMap[key]
		base, perr := parseDecimalToInt(strings.TrimSpace(entry.Base))
		if perr != nil {
			problems = append(problems, fmt.Sprintf("Entry key %q has invalid base %q: %v", key, entry.Base, perr))
			convertedRoots = append(convertedRoots, big.NewInt(0)) // 👈 put zero instead of skipping
			continue
		}
		if base < 2 || base > 16 {
			problems = append(problems, fmt.Sprintf("Entry key %q has unsupported base %d (allowed 2..16).", key, base))
			convertedRoots = append(convertedRoots, big.NewInt(0)) // 👈 put zero instead of skipping
			continue
		}
		valBig, verr := convertValueToBig(entry.Value, base)
		if verr != nil {
			problems = append(problems, fmt.Sprintf("Entry key %q with base %d and value %q is invalid: %v", key, base, entry.Value, verr))
			convertedRoots = append(convertedRoots, big.NewInt(0)) // 👈 put zero instead of skipping
			continue
		}
		convertedRoots = append(convertedRoots, valBig)
	}

	if len(problems) > 0 {
		fmt.Println("Wrong Dataset")
		for _, p := range problems {
			fmt.Println("-", p)
		}
		return
	}

	if len(convertedRoots) != n {
		fmt.Println("Wrong Dataset")
		fmt.Printf("- After conversions, number of valid roots (%d) differs from 'n' (%d).\n", len(convertedRoots), n)
		return
	}
	selected := make([]*big.Int, 0, k)
	for i := 0; i < k; i++ {
		selected = append(selected, new(big.Int).Set(convertedRoots[i]))
	}

	coeffs := []*big.Int{big.NewInt(1)}
	for _, r := range selected {
		newCoeffs := make([]*big.Int, len(coeffs)+1)
		for i := range newCoeffs {
			newCoeffs[i] = big.NewInt(0)
		}
		for i := 0; i < len(coeffs); i++ {
			newCoeffs[i+1].Add(newCoeffs[i+1], new(big.Int).Set(coeffs[i]))
			temp := new(big.Int).Mul(coeffs[i], r)
			temp.Neg(temp)
			newCoeffs[i].Add(newCoeffs[i], temp)
		}
		coeffs = newCoeffs
	}

	out := make([]string, 0, len(coeffs))
	for i := len(coeffs) - 1; i >= 0; i-- {
		out = append(out, coeffs[i].String())
	}
	fmt.Println(strings.Join(out, " "))
}
