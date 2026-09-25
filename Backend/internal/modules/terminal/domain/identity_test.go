package domain

import "testing"

func TestResourceIdentityIsDeterministic(t *testing.T) {
	containerA, volumeA, err := ResourceIdentity("project-1")
	if err != nil {
		t.Fatal(err)
	}
	containerB, volumeB, err := ResourceIdentity("project-1")
	if err != nil {
		t.Fatal(err)
	}
	if containerA != containerB || volumeA != volumeB {
		t.Fatalf("identity is not deterministic: %q/%q vs %q/%q", containerA, volumeA, containerB, volumeB)
	}
}
