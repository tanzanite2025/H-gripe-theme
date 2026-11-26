package product

import "testing"

func TestIsRetiredWheelsetSpecificationTemplateSlug(t *testing.T) {
	for _, slug := range []string{"wheelset", "WHEELSET", " wheelset "} {
		if !IsRetiredWheelsetSpecificationTemplateSlug(slug) {
			t.Errorf("expected %q to be a retired wheelset template slug", slug)
		}
	}
	for _, slug := range []string{"wheelsets", "wheelset_options", "rim"} {
		if IsRetiredWheelsetSpecificationTemplateSlug(slug) {
			t.Errorf("did not expect %q to be a retired wheelset template slug", slug)
		}
	}
}
