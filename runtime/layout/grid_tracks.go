package layout

import (
	"strconv"
	"strings"
)

// gridTrack is one parsed track spec entry.
type gridTrack struct {
	fixed  int  // resolved size for cell/% tracks
	fr     int  // fr weight (0 = fixed)
	minFix int  // minmax minimum, enforced after fr distribution
	auto   bool // auto track (treated as 1fr)
}

// parseTrackList parses a grid-template-columns/rows value into track sizes
// resolved against the available axis size. Supports cell/ch/% lengths, fr,
// auto, repeat(n, ...), and minmax(min, max). Per the svelterm deviation,
// minmax minimums are enforced without redistribution.
func parseTrackList(spec string, avail, gap int) []int {
	tokens := expandRepeat(tokenizeTracks(spec))
	tracks := make([]gridTrack, 0, len(tokens))
	for _, tok := range tokens {
		tracks = append(tracks, parseTrack(tok, avail))
	}
	return resolveTracks(tracks, avail, gap)
}

// tokenizeTracks splits on top-level spaces (paren-aware).
func tokenizeTracks(spec string) []string {
	var tokens []string
	depth, start := 0, -1
	for i := 0; i < len(spec); i++ {
		switch spec[i] {
		case '(':
			depth++
		case ')':
			depth--
		case ' ':
			if depth == 0 {
				if start >= 0 {
					tokens = append(tokens, spec[start:i])
					start = -1
				}
				continue
			}
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		tokens = append(tokens, spec[start:])
	}
	return tokens
}

// expandRepeat expands repeat(n, tracks...) tokens in place.
func expandRepeat(tokens []string) []string {
	var out []string
	for _, tok := range tokens {
		if !strings.HasPrefix(tok, "repeat(") || !strings.HasSuffix(tok, ")") {
			out = append(out, tok)
			continue
		}
		body := tok[len("repeat(") : len(tok)-1]
		count, rest, found := strings.Cut(body, ",")
		if !found {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(count))
		if err != nil || n <= 0 {
			continue
		}
		sub := tokenizeTracks(strings.TrimSpace(rest))
		for i := 0; i < n; i++ {
			out = append(out, sub...)
		}
	}
	return out
}

func parseTrack(tok string, avail int) gridTrack {
	if strings.HasPrefix(tok, "minmax(") && strings.HasSuffix(tok, ")") {
		body := tok[len("minmax(") : len(tok)-1]
		minTok, maxTok, found := strings.Cut(body, ",")
		if !found {
			return gridTrack{}
		}
		t := parseTrack(strings.TrimSpace(maxTok), avail)
		min := parseTrack(strings.TrimSpace(minTok), avail)
		t.minFix = min.fixed
		return t
	}
	if tok == "auto" {
		return gridTrack{auto: true, fr: 1}
	}
	if strings.HasSuffix(tok, "fr") {
		n, err := strconv.Atoi(strings.TrimSuffix(tok, "fr"))
		if err != nil || n <= 0 {
			n = 1
		}
		return gridTrack{fr: n}
	}
	if strings.HasSuffix(tok, "%") {
		n, err := strconv.Atoi(strings.TrimSuffix(tok, "%"))
		if err != nil {
			return gridTrack{}
		}
		return gridTrack{fixed: avail * n / 100}
	}
	return gridTrack{fixed: ParseCellLength(tok)}
}

// resolveTracks distributes the axis size: fixed tracks first, then fr
// weights share the remainder (after gaps); minmax minimums clamp last.
func resolveTracks(tracks []gridTrack, avail, gap int) []int {
	if len(tracks) == 0 {
		return nil
	}
	gaps := gap * (len(tracks) - 1)
	fixedTotal, frTotal := 0, 0
	for _, t := range tracks {
		fixedTotal += t.fixed
		frTotal += t.fr
	}
	remaining := avail - fixedTotal - gaps
	if remaining < 0 {
		remaining = 0
	}
	sizes := make([]int, len(tracks))
	used := 0
	frSeen := 0
	for i, t := range tracks {
		if t.fr == 0 {
			sizes[i] = t.fixed
			continue
		}
		frSeen += t.fr
		// Cumulative rounding keeps the total exact.
		share := remaining*frSeen/frTotal - used
		used += share
		sizes[i] = share
	}
	for i, t := range tracks {
		if t.minFix > 0 && sizes[i] < t.minFix {
			sizes[i] = t.minFix // enforced without redistribution
		}
	}
	return sizes
}
