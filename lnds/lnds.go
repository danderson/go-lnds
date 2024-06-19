// Package lnds computes the Longest Non-Decreasing Subsequence (LNDS)
// of a slice of comparable items. Put simply: given a list, this
// package tells you which elements have to be removed, in order for
// the remaining shorter list to be correctly sorted.
//
// LNDS is a minor extension of the better known Longest Increasing
// Subsequence (LIS) algorithm. The only difference is that increasing
// subsequences cannot contain equal elements, whereas nondecreasing
// subsequences can.
//
// Increasing and nondecreasing subsequence algorithms are also
// closely related to sorting algorithms. You could think of LNDS as a
// sort function that instead of sorting the input just tells you
// which elements need to move elsewhere in the list.
//
// LNDS's time complexity is comparable to that of sorting algorithms:
// for an input of length n, it takes O(n·logn) time in the worst
// case, Θ(n) for the best case of an already sorted list, and
// O(n·logn) in the average case.
//
// The exact implementation is not guaranteed to remain the same, but
// at present it uses the algorithm discovered by Fredman [1] and
// Knuth [2, Algorithm I]. At its core, it's the Schensted insertion
// algorithm [3], stripped down and with additional optimizations
// suitable for execution by a computer, rather than a mathematician.
//
// The references above specify an algorithm for longest increasing
// subsequence, and only compute the length of the subsequence. This
// package makes the minor extensions necessary to support equal
// elements, and to track enough state to produce the full
// subsequence. These are not original extensions, and in fact are
// ubiquitous around the internet and in textbooks, but I've been
// unable to track down their originator.
//
// [1]: Michael L. Fredman, "On computing the length of longest
//      increasing subsequences", Discrete Mathematics, vol. 54, issue
//      1, pp. 29-35. Available: https://doi.org/10.1016/0012-365X(75)90103-X
// [2]: Donald E. Knuth, "The Art of Computer Programming", vol. 3,
//      section 5.1.4, Algorithm I
// [3]: Craige Schensted, “Longest Increasing and Decreasing
//      Subsequences,” Canadian Journal of Mathematics, vol. 13,
//      pp. 179–191, 1961. Available: https://doi:10.4153/CJM-1961-015-3
package lnds

import "slices"

func PartitionUnsorted[E any](lst []E, cmp func(E, E) int) (sorted, notSorted []E) {
	st := state[E]{
		cmp:   cmp,
		elts:  lst,
		tails: make([]int, 1, len(lst)),
		prev:  make([]int, len(lst)),
	}
	return st.run()
}

type state[E any] struct {
	cmp   func(E, E) int
	elts  []E
	tails []int
	prev  []int
}

func (st *state[E]) run() (sorted, unsorted []E) {
	for i := range st.elts {
		st.process(i)
	}

	if len(st.tails) == len(st.elts) {
		// Input was already sorted.
		return st.elts, nil
	}

	return st.result()
}

func (st *state[E]) process(newIdx int) {
	if newIdx == 0 {
		st.prev[0] = -1
		st.tails[0] = 0
		return
	}

	curBestIdx := st.tails[len(st.tails)-1]
	if st.cmp(st.elts[newIdx], st.elts[curBestIdx]) >= 0 {
		// Fast path: can extend the current longest sequence.
		st.prev[newIdx] = curBestIdx
		st.tails = append(st.tails, newIdx)
		return
	}

	replaceIdx, found := slices.BinarySearchFunc(st.tails, newIdx, st.eltIdxCmp)
	if found {
		// TODO: make my own binary search instead, that selects the
		// largest equal element if found=true.
		for {
			replaceIdx++
			switch st.eltIdxCmp(replaceIdx, newIdx) {
			case 0:
				continue
			case +1:
				// newIdx is a better tail for this length subsequence.
				break
			case -1:
				// existing idx is a better tail, newIdx is not part of a
				// worthy subsequence.
				return
			}
		}
	}
	if replaceIdx == 0 {
		st.prev[newIdx] = -1
	} else {
		st.prev[newIdx] = st.tails[replaceIdx-1]
	}
	st.tails[replaceIdx] = newIdx
}

func (st *state[E]) eltIdxCmp(a, b int) int {
	return st.cmp(st.elts[a], st.elts[b])
}

func (st *state[E]) result() (sorted, unsorted []E) {
	sorted = make([]E, len(st.tails))
	unsorted = make([]E, len(st.elts)-len(st.tails))

	seqIdx := st.tails[len(st.tails)-1] // current element of LIS
	allIdx := len(st.elts) - 1          // current input element
	sortedIdx := len(sorted) - 1        // next write position into sorted
	unsortedIdx := len(unsorted) - 1    // next write position into unsorted

	for allIdx >= 0 {
		for allIdx >= 0 && seqIdx == allIdx {
			sorted[sortedIdx] = st.elts[seqIdx]
			seqIdx = st.prev[seqIdx]
			allIdx--
			sortedIdx--
		}

		// Did we fall out of the previous loop because seqIdx jumped
		// away from allIdx? That means there are unsorted elements to
		// output.
		for seqIdx < allIdx {
			unsorted[unsortedIdx] = st.elts[allIdx]
			allIdx--
			unsortedIdx--
		}
	}

	return sorted, unsorted
}
