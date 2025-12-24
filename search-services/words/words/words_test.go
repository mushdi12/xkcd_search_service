package words

import "testing"

func TestNorm_RemovesStopWordsAndNormalizes(t *testing.T) {
	phrase := "An Apple a day keeps Doctors away!"

	result := Norm(phrase)

	if len(result) == 0 {
		t.Fatalf("expected some words, got 0")
	}

	found := false
	for _, w := range result {
		if w == "appl" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected stemmed word \"appl\" in result: %#v", result)
	}
}
