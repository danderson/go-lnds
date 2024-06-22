package lis

import (
	"cmp"
	"math/rand"
	"slices"
	"testing"

	"github.com/creachadair/mds/slice"
	diff "github.com/google/go-cmp/cmp"
)

func TestLIS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		in         []int
		wantSorted []int
		wantRest   []int
	}{
		{
			name: "nil",
		},
		{
			name: "empty",
			in:   []int{},
		},
		{
			name:       "singleton",
			in:         []int{1},
			wantSorted: []int{1},
			wantRest:   []int{},
		},
		{
			name:       "sorted",
			in:         []int{1, 2, 3, 4},
			wantSorted: []int{1, 2, 3, 4},
			wantRest:   []int{},
		},
		{
			name:       "backwards",
			in:         []int{4, 3, 2, 1},
			wantSorted: []int{1},
			wantRest:   []int{4, 3, 2},
		},
		{
			name:       "organ_pipe",
			in:         []int{1, 2, 3, 4, 3, 2, 1},
			wantSorted: []int{1, 2, 3, 3},
			wantRest:   []int{4, 2, 1},
		},
		{
			name:       "sawtooth",
			in:         []int{0, 1, 0, -1, 0, 1, 0, -1},
			wantSorted: []int{0, 0, 0, 0},
			wantRest:   []int{1, -1, 1, -1},
		},
		{
			name:       "A005132", // from oeis.org
			in:         []int{0, 1, 3, 6, 2, 7, 13, 20, 12, 21, 11, 22, 10},
			wantSorted: []int{0, 1, 3, 6, 7, 13, 20, 21, 22},
			wantRest:   []int{2, 12, 11, 10},
		},
		{
			name:       "swapped_pairs",
			in:         []int{2, 1, 4, 3, 6, 5, 8, 7},
			wantSorted: []int{1, 3, 5, 7},
			wantRest:   []int{2, 4, 6, 8},
		},
		{
			name: "run_of_equals",
			// swapped_pairs with more 3s sprinkled in.
			in:         []int{2, 1, 3, 4, 3, 6, 3, 5, 8, 3, 7},
			wantSorted: []int{1, 3, 3, 3, 3, 7},
			wantRest:   []int{2, 4, 6, 5, 8},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotSorted, gotRest := LIS(tc.in, cmp.Compare)
			if diff := diff.Diff(gotSorted, tc.wantSorted); diff != "" {
				t.Errorf("LIS subsequence is wrong (-got+want):\n%s", diff)
			}
			if diff := diff.Diff(gotRest, tc.wantRest); diff != "" {
				t.Errorf("LIS remainder is wrong (-got+want):\n%s", diff)
			}
			if t.Failed() {
				t.Logf("Input was: %v", tc.in)
				t.Logf("Got: %v, %v", gotSorted, gotRest)
				t.Logf("Want: %v, %v", tc.wantSorted, tc.wantRest)
			}
		})
	}
}

func TestLISAgainstLCS(t *testing.T) {
	t.Parallel()

	// A result from literature relates LIS and LCS:
	//
	//   len(LIS(lst)) == len(LCS(lst, Sorted(lst)))
	//
	// Check that this holds true. Ideally we could also compare the
	// actual resultant lists, but there's no guarantee that LIS and
	// LCS will return the _same_ longest increasing subsequence, if
	// multiple options are available.

	const numVals = 50
	const numIters = 100
	for i := 0; i < numIters; i++ {
		input := randomInts(numVals)

		gotLIS, _ := LIS(input, cmp.Compare)

		sorted := append([]int(nil), input...)
		slices.Sort(sorted)
		gotLCS := slice.LCS(input, sorted)

		if got, want := len(gotLIS), len(gotLCS); got != want {
			t.Logf("Input: %v", input)
			t.Errorf("len(LIS(x)) = %v, want len(LCS(x, sorted(x))) = %v", got, want)
		}
	}
}

func TestLISRandom(t *testing.T) {
	t.Parallel()

	const numVals = 50
	const numIters = 100

	for i := 0; i < numIters; i++ {
		input := randomInts(numVals)
		wantSorted := quadraticLIS(input)
		gotSorted, _ := LIS(input, cmp.Compare)

		if diff := diff.Diff(gotSorted, wantSorted); diff != "" {
			t.Logf("Input: %v", input)
			t.Errorf("LIS subsequence is wrong (-got+want):\n%s", diff)
		}
	}
}

// quadraticLIS returns the same longest increasing subsequence of lst
// that LIS() returns, but using a quadratic recursive search that is
// much slower, but more obviously correct by inspection.
func quadraticLIS(lst []int) []int {
	// cmpSet orders a and b according to the best LIS. Longest lists
	// go first, and within that equivalence class lists with smaller
	// elements go first.
	cmpSeq := func(a, b []int) int {
		if res := cmp.Compare(len(a), len(b)); res != 0 {
			return -res
		}
		for i := range a {
			if res := cmp.Compare(a[i], b[i]); res != 0 {
				return res
			}
		}
		// fully equal, which can happen in the quadratic algorithm
		// since we might generate permutations of indistinguishable
		// equal elements.
		return 0
	}

	// findLIS recursively constructs all possible increasing
	// sequences of vs, updating best as it discovers better LIS
	// candidates.
	var findLIS func([]int, []int, []int) []int
	findLIS = func(vs, acc, best []int) (bestOfTree []int) {
		if len(vs) == 0 {
			if cmpSeq(acc, best) < 0 {
				best = append(best[:0], acc...)
			}
			return best
		}

		lnBest := len(best)
		if lnBest > 0 && len(vs)+len(acc) < lnBest {
			// can't possibly do better than what's already known,
			// give up early.
			return best
		}

		elt, vs := vs[0], vs[1:]
		if len(acc) == 0 || elt >= acc[len(acc)-1] {
			// elt could extend acc, try that
			best = findLIS(vs, append(acc, elt), best)
		}
		// and always try skipping elt
		return findLIS(vs, acc, best)
	}

	// Preallocate, so the recursion doesn't add insult to injury by
	// allocating as well.
	acc := make([]int, 0, len(lst))
	best := make([]int, 0, len(lst))

	return findLIS(lst, acc, best)
}

func randomInts(N int) []int {
	ret := make([]int, N)
	for i := range ret {
		ret[i] = rand.Intn(2 * N)
	}
	return ret
}
