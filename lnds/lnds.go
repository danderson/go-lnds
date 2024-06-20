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
// Knuth [2]. At its core, it's the Schensted insertion
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
// [1]: Michael L. Fredman, "On computing the length of longest increasing subsequences", Discrete Mathematics, vol. 54, issue 1, pp. 29-35. Available: https://doi.org/10.1016/0012-365X(75)90103-X
// [2]: Donald E. Knuth, "The Art of Computer Programming", vol. 3, section 5.1.4, Algorithm I
// [3]: Craige Schensted, “Longest Increasing and Decreasing Subsequences,” Canadian Journal of Mathematics, vol. 13, pp. 179–191, 1961. Available: https://doi:10.4153/CJM-1961-015-3
package lnds

import "slices"

// LNDS computes a longest non-decreasing subsequence of vs, whose
// elements must be totally ordered by cmp.
func LNDS[T any, Slice ~[]T](lst Slice, cmp func(T, T) int) (sorted, rest Slice) {
	// Editorial note: "longest non-decreasing subsequence" is a
	// mouthful, so the comments in this function omit
	// "non-decreasing" and just say "subsequence" or "longest
	// subsequence". This is unambiguous since we don't handle any
	// subsequences other than the non-decreasing kind.
	//
	// The algorithm's core is a loop over every element of lst. Each
	// element we consider can be the beginning of a new subsequence
	// of length 1, and may be a valid extension of some previously
	// found subsequences. Conceptually, the loop is constructing this
	// vast tree of all possible valid subsequences, and then picks
	// one of the longest to return.
	//
	// However, this would be quite inefficient. Thankfully there are
	// four optimizations we can apply. Leaving proofs of their
	// correctness to the literature, they are:
	//
	//   - Each element only gets to participate in creating the
	//     longest subsequences it can, we discard all shorter
	//     options. It might still participate in many subsequences of
	//     that length, but that brings us to...
	//   - We only need to remember one subsequence of every length,
	//     the one whose final element is the smallest. This is the
	//     tails array, and means that every new element will
	//     contribute to exactly 0 or 1 subsequence of interest.
	//   - Elements always appear in the tails array in sorted order,
	//     so we can use a binary search to find the one subsequence
	//     that a new element might contribute to. This gets us
	//     O(log(n)) time per element, instead of O(n).
	//   - Successive non-decreasing new elements always contribute to
	//     the longest currently known subsequence, which is the final
	//     entry of tails. If we check for this trivial case before
	//     embarking on the binary search, elements that appear in
	//     non-decreasing order can be processed in O(1) time rather
	//     than O(log(n)).

	var (
		// tails[L] is the index into lst for the final element of a
		// subsequence of length L. If several such subsequences
		// exist, tails keeps whichever has the smallest final
		// element, according to cmp.
		tails = make([]int, 1, len(lst))

		// prev[i] is the index into lst for the element that comes
		// before lst[i] in a subsequence tracked by tails, or -1 if
		// lst[i] is the first in a subsequence. It's effectively the
		// pointers for the linked lists whose first elements are
		// tracked in tails.
		//
		// tails by itself only gives us the length of the longest
		// subsequence, and its final element. prev is the additional
		// state we need to reconstruct the entire subsequence.
		//
		// prev[i]'s value is only valid if lst[i] is part of a
		// subsequence currently being tracked in tails.
		prev = make([]int, len(lst))
	)

processElement:
	for i := range lst {
		if i == 0 {
			// The rest of this loop is cleaner if it can assume that
			// i-1 exists. This handles the initial edge case.
			prev[0] = -1
			tails[0] = 0
			continue
		}

		idxOfBestTail := len(tails) - 1
		if cmp(lst[i], lst[idxOfBestTail]) >= 0 {
			// Fast path: the i-th element extends the currently known
			// longest subsequence.
			prev[i] = idxOfBestTail
			tails = append(tails, i)
			continue
		}

		// Otherwise, the i-th element can only produce a shorter
		// subsequence. Figure out what length, and whether this new
		// subsequence is better than the one tails already knew
		// about.
		//
		// TODO: a custom BinarySearch implementation could bias
		// towards the last matching element, rather than the first.
		replaceIdx, found := slices.BinarySearchFunc(tails[:len(tails)-1], i, func(i, j int) int {
			return cmp(lst[i], lst[j])
		})
		if found {
			// lst has equal elements, and we've just found one. In a
			// non-decreasing subsequence, we can chain the equal
			// elements together, but slices.BinarySearchFunc gave us
			// the index of the _first_ occurrence of the equal
			// element. Scan forward to go one past the _last_
			// occurrence.
			for {
				replaceIdx++
				switch cmp(lst[replaceIdx], lst[i]) {
				case 0:
					continue
				case +1:
					break // new element is better than what tails has
				case -1:
					continue processElement // new element is worse than what tails has
				}
			}
		}
		// The new element is extending the subsequence tracked in
		// replaceIdx-1, replacing the previous best extension that
		// was stored in replaceIdx. We have to deal with the edge
		// case of the single-element subsequence.
		if replaceIdx == 0 {
			prev[i] = -1
		} else {
			prev[i] = tails[replaceIdx-1]
		}
		tails[replaceIdx] = i
	}

	// We can now iterate back through the longest subsequence and
	// partition the input.
	sorted = make([]E, len(tails))
	rest = make([]E, len(lst)-len(tails))
	var (
		seqIdx    = tails[len(tails)-1] // current longest subsequence element
		allIdx    = len(lst) - 1        // current input element
		sortedIdx = len(sorted) - 1
		restIdx   = len(rest) - 1
	)
output:
	for {
		for seqIdx == allIdx {
			sorted[sortedIdx] = lst[seqIdx]
			seqIdx = prev[seqIdx]
			allIdx--
			sortedIdx--

			if allIdx < 0 {
				break output
			}
		}

		// We fell out of the previous loop because seqIdx jumped
		// ahead of allIdx, indicating one or more elements that
		// aren't part of the longest subsequence.
		for seqIdx < allIdx {
			rest[restIdx] = lst[allIdx]
			allIdx--
			restIdx--

			if allIdx < 0 {
				break output
			}
		}
	}

	return sorted, rest
}
