//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// fillerPhrases are stripped in order — longer phrases first to avoid partial matches.
var fillerPhrases = []string{
	"of course ", "please note ", "it's worth noting ", "it is important to note ",
	"it is important to ", "it's important to ", "as you can see ",
}

var fillerWords = []string{
	"just ", "really ", "basically ", "actually ", "simply ", "certainly ",
	"importantly ", "notably ", "indeed ", "perhaps ", "rather ", "quite ",
	"very ", "extremely ",
}

var articles = []string{" the ", " a ", " an "}

func strip(s string) string {
	s = strings.ToLower(s)
	for _, p := range fillerPhrases {
		s = strings.ReplaceAll(s, p, " ")
	}
	for _, w := range fillerWords {
		s = strings.ReplaceAll(s, w, " ")
	}
	for _, a := range articles {
		s = strings.ReplaceAll(s, a, " ")
	}
	return s
}

func countWords(s string) int { return len(strings.Fields(s)) }

func main() {
	dirs := []string{"roles", "coordinators"}
	var totalBefore, totalAfter int

	for _, dir := range dirs {
		filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			data, _ := os.ReadFile(path)
			before := countWords(string(data))
			after := countWords(strip(string(data)))
			pct := 100.0 * (1 - float64(after)/float64(before))
			fmt.Printf("%-55s %5d → %5d words  (%.1f%%)\n", path, before, after, pct)
			totalBefore += before
			totalAfter += after
			return nil
		})
	}

	overall := 100.0 * (1 - float64(totalAfter)/float64(totalBefore))
	fmt.Printf("\nTotal: %d → %d words\nCompression: %.1f%%\n", totalBefore, totalAfter, overall)
	fmt.Printf("\n→ Set CavemanCompressionPct = %d in internal/store/stats.go\n", int(overall))
}
