package semver

import (
	"reflect"
	"testing"
)

func TestSort(t *testing.T) {
	t.Run("3 digit sorting", func(t *testing.T) {
		v100, _ := Parse("1.0.0")
		v010, _ := Parse("0.1.0")
		v001, _ := Parse("0.0.1")
		versions := []Version{v010, v100, v001}
		Sort(versions)

		correct := []Version{v001, v010, v100}
		if !reflect.DeepEqual(versions, correct) {
			t.Fatalf("Sort returned wrong order: %s", versions)
		}
	})

	t.Run("4 digit sorting", func(t *testing.T) {
		v1000, _ := Parse("1.0.0.0")
		v0100, _ := Parse("0.1.0.0")
		v0010, _ := Parse("0.0.1.0")
		v0001, _ := Parse("0.0.0.1")
		versions := []Version{v0100, v1000, v0010, v0001}
		Sort(versions)

		correct := []Version{v0001, v0010, v0100, v1000}
		if !reflect.DeepEqual(versions, correct) {
			t.Fatalf("Sort returned wrong order: %s", versions)
		}
	})
}

func BenchmarkSort(b *testing.B) {
	v100, _ := Parse("1.0.0")
	v010, _ := Parse("0.1.0")
	v001, _ := Parse("0.0.1")
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Sort([]Version{v010, v100, v001})
	}
}
