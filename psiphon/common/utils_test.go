/*
 * Copyright (c) 2014, Psiphon Inc.
 * All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"
)

func TestGetStringSlice(t *testing.T) {

	originalSlice := []string{"a", "b", "c"}

	j, err := json.Marshal(originalSlice)
	if err != nil {
		t.Errorf("json.Marshal failed: %s", err)
	}

	var value interface{}

	err = json.Unmarshal(j, &value)
	if err != nil {
		t.Errorf("json.Unmarshal failed: %s", err)
	}

	newSlice, ok := GetStringSlice(value)
	if !ok {
		t.Errorf("GetStringSlice failed")
	}

	if !reflect.DeepEqual(originalSlice, newSlice) {
		t.Errorf("unexpected GetStringSlice output")
	}
}

func TestMakeSecureRandomPerm(t *testing.T) {
	for n := 0; n < 1000; n++ {
		perm, err := MakeSecureRandomPerm(n)
		if err != nil {
			t.Errorf("MakeSecureRandomPerm failed: %s", err)
		}
		if len(perm) != n {
			t.Error("unexpected permutation size")
		}
		sum := 0
		for i := 0; i < n; i++ {
			sum += perm[i]
		}
		expectedSum := (n * (n - 1)) / 2
		if sum != expectedSum {
			t.Error("unexpected permutation")
		}
	}
}

func TestMakeSecureRandomRange(t *testing.T) {
	min := 1
	max := 19
	var gotMin, gotMax bool
	for n := 0; n < 1000; n++ {
		i, err := MakeSecureRandomRange(min, max)
		if err != nil {
			t.Errorf("MakeSecureRandomRange failed: %s", err)
		}
		if i < min || i > max {
			t.Error("out of range")
		}
		if i == min {
			gotMin = true
		}
		if i == max {
			gotMax = true
		}
	}
	if !gotMin {
		t.Error("missing min")
	}
	if !gotMax {
		t.Error("missing max")
	}
}

func TestMakeSecureRandomPeriod(t *testing.T) {
	min := 1 * time.Nanosecond
	max := 10000 * time.Nanosecond

	different := 0

	for n := 0; n < 1000; n++ {
		res1, err := MakeSecureRandomPeriod(min, max)

		if err != nil {
			t.Errorf("MakeSecureRandomPeriod failed: %s", err)
		}

		if res1 < min {
			t.Error("duration should not be less than min")
		}

		if res1 > max {
			t.Error("duration should not be more than max")
		}

		res2, err := MakeSecureRandomPeriod(min, max)

		if err != nil {
			t.Errorf("MakeSecureRandomPeriod failed: %s", err)
		}

		if res1 != res2 {
			different += 1
		}
	}

	// res1 and res2 should be different most of the time, but it's possible
	// to get the same result twice in a row.
	if different < 900 {
		t.Error("duration insufficiently random")
	}
}

func TestJitter(t *testing.T) {

	testCases := []struct {
		n           int64
		factor      float64
		expectedMin int64
		expectedMax int64
	}{
		{100, 0.1, 90, 110},
		{1000, 0.3, 700, 1300},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("jitter case: %+v", testCase), func(t *testing.T) {

			min := int64(math.MaxInt64)
			max := int64(0)

			for i := 0; i < 100000; i++ {

				x := Jitter(testCase.n, testCase.factor)
				if x < min {
					min = x
				}
				if x > max {
					max = x
				}
			}

			if min != testCase.expectedMin {
				t.Errorf("unexpected minimum jittered value: %d", min)
			}

			if max != testCase.expectedMax {
				t.Errorf("unexpected maximum jittered value: %d", max)
			}
		})
	}
}

func TestCompress(t *testing.T) {

	originalData := []byte("test data")

	compressedData := Compress(originalData)

	decompressedData, err := Decompress(compressedData)
	if err != nil {
		t.Errorf("Uncompress failed: %s", err)
	}

	if bytes.Compare(originalData, decompressedData) != 0 {
		t.Error("decompressed data doesn't match original data")
	}
}

func TestFormatByteCount(t *testing.T) {

	testCases := []struct {
		n              uint64
		expectedOutput string
	}{
		{500, "500B"},
		{1024, "1.0K"},
		{10000, "9.8K"},
		{1024*1024 + 1, "1.0M"},
		{100*1024*1024 + 99999, "100.1M"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.expectedOutput, func(t *testing.T) {
			output := FormatByteCount(testCase.n)
			if output != testCase.expectedOutput {
				t.Errorf("unexpected output: %s", output)
			}
		})
	}
}

func TestWeightedCoinFlip(t *testing.T) {

	runs := 100000
	tolerance := 1000

	testCases := []struct {
		weight        float64
		expectedTrues int
	}{
		{0.333, runs / 3},
		{0.5, runs / 2},
		{1.0, runs},
		{0.0, 0},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%f", testCase.weight), func(t *testing.T) {
			trues := 0
			for i := 0; i < runs; i++ {
				if FlipWeightedCoin(testCase.weight) {
					trues++
				}
			}

			min := testCase.expectedTrues - tolerance
			if min < 0 {
				min = 0
			}
			max := testCase.expectedTrues + tolerance

			if trues < min || trues > max {
				t.Errorf("unexpected coin flip outcome: %f %d (+/-%d) %d",
					testCase.weight, testCase.expectedTrues, tolerance, trues)
			}
		})
	}
}
