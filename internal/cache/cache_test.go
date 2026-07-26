package cache

import (
	"testing"
)

func TestHashDiff(t *testing.T) {
	diff1 := "diff --git a/main.go b/main.go\n+func main() {}"
	diff2 := "diff --git a/main.go b/main.go\n+func main() {}"
	diff3 := "diff --git a/main.go b/main.go\n+func test() {}"

	hash1 := ComputeHash(diff1)
	hash2 := ComputeHash(diff2)
	hash3 := ComputeHash(diff3)

	if hash1 != hash2 {
		t.Errorf("HashDiff() should produce identical hash for identical diffs. Got %s and %s", hash1, hash2)
	}

	if hash1 == hash3 {
		t.Errorf("HashDiff() produced matching hash for different diffs")
	}
}
