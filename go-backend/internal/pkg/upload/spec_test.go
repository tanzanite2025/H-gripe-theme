package upload

import (
	"bytes"
	"image"
	"image/png"
	"mime/multipart"
	"sort"
	"testing"
)

func TestListUploadSpecsUsesStableCodeOrder(t *testing.T) {
	specs := ListUploadSpecs()
	if len(specs) != len(uploadSpecDefinitions) {
		t.Fatalf("ListUploadSpecs returned %d specs, want %d", len(specs), len(uploadSpecDefinitions))
	}

	codes := make([]string, 0, len(specs))
	for _, spec := range specs {
		codes = append(codes, spec.Code)
	}
	if !sort.StringsAreSorted(codes) {
		t.Fatalf("upload specs are not sorted by code: %v", codes)
	}

	product, ok := GetUploadSpec("  PRODUCT_IMAGE ")
	if !ok {
		t.Fatal("GetUploadSpec did not normalize the spec code")
	}
	if product.RecommendedWidth != 1600 || product.RecommendedHeight != 1600 {
		t.Fatalf("product image recommendation = %dx%d, want 1600x1600", product.RecommendedWidth, product.RecommendedHeight)
	}

	product.AcceptedExtensions[0] = ".mutated"
	fresh, ok := GetUploadSpec(string(SpecProductImage))
	if !ok || fresh.AcceptedExtensions[0] == ".mutated" {
		t.Fatal("GetUploadSpec returned mutable registry data")
	}
}

func TestUploadSpecRegistryContainsCriticalImageContracts(t *testing.T) {
	tests := []struct {
		code                SpecCode
		recommendedWidth    int
		recommendedHeight   int
		recommendedLongEdge int
		exactWidth          int
		exactHeight         int
		aspectRatioLabel    string
		maxFiles            int
	}{
		{
			code:              SpecFAQAnswerImage,
			recommendedWidth:  800,
			recommendedHeight: 800,
			exactWidth:        800,
			exactHeight:       800,
		},
		{
			code:              SpecVisualShowcaseHomeCategories,
			recommendedWidth:  1920,
			recommendedHeight: 1080,
			aspectRatioLabel:  "16:9",
		},
		{
			code:              SpecVisualShowcaseEditorial,
			recommendedWidth:  1200,
			recommendedHeight: 1600,
			aspectRatioLabel:  "3:4",
		},
		{
			code:                SpecMediaLibraryImage,
			recommendedLongEdge: 1600,
		},
		{
			code:                SpecUserShowcaseImage,
			recommendedLongEdge: 1600,
			maxFiles:            10,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			spec, ok := GetUploadSpec(string(tt.code))
			if !ok {
				t.Fatalf("missing upload spec %q", tt.code)
			}
			if spec.RecommendedWidth != tt.recommendedWidth || spec.RecommendedHeight != tt.recommendedHeight {
				t.Fatalf(
					"recommendation = %dx%d, want %dx%d",
					spec.RecommendedWidth,
					spec.RecommendedHeight,
					tt.recommendedWidth,
					tt.recommendedHeight,
				)
			}
			if spec.RecommendedLongEdge != tt.recommendedLongEdge {
				t.Fatalf("recommended long edge = %d, want %d", spec.RecommendedLongEdge, tt.recommendedLongEdge)
			}
			if spec.ExactWidth != tt.exactWidth || spec.ExactHeight != tt.exactHeight {
				t.Fatalf("exact dimensions = %dx%d, want %dx%d", spec.ExactWidth, spec.ExactHeight, tt.exactWidth, tt.exactHeight)
			}
			if spec.AspectRatioLabel != tt.aspectRatioLabel {
				t.Fatalf("aspect ratio = %q, want %q", spec.AspectRatioLabel, tt.aspectRatioLabel)
			}
			if spec.MaxFiles != tt.maxFiles {
				t.Fatalf("max files = %d, want %d", spec.MaxFiles, tt.maxFiles)
			}
		})
	}
}

func TestValidateSpecFileEnforcesVisualAspectRatios(t *testing.T) {
	tests := []struct {
		name    string
		code    SpecCode
		width   int
		height  int
		wantErr bool
	}{
		{
			name:   "home visual accepts 16 by 9",
			code:   SpecVisualShowcaseHomeCategories,
			width:  160,
			height: 90,
		},
		{
			name:    "home visual rejects non 16 by 9",
			code:    SpecVisualShowcaseHomeCategories,
			width:   160,
			height:  100,
			wantErr: true,
		},
		{
			name:   "editorial visual accepts 3 by 4",
			code:   SpecVisualShowcaseEditorial,
			width:  120,
			height: 160,
		},
		{
			name:    "editorial visual rejects non 3 by 4",
			code:    SpecVisualShowcaseEditorial,
			width:   120,
			height:  150,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSpecFile(pngFileHeader(t, tt.name+".png", tt.width, tt.height), string(tt.code))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected aspect ratio validation to fail")
				}
				if ErrorCode(err) != CodeInvalidDimensions {
					t.Fatalf("error code = %q, want %q", ErrorCode(err), CodeInvalidDimensions)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected aspect ratio validation to pass, got %v", err)
			}
		})
	}
}

func TestValidateSpecFileRejectsUnknownSpec(t *testing.T) {
	err := ValidateSpecFile(nil, "unknown_image_purpose")
	if err == nil {
		t.Fatal("expected unknown upload spec to be rejected")
	}
	if ErrorCode(err) != CodeInvalidType {
		t.Fatalf("error code = %q, want %q", ErrorCode(err), CodeInvalidType)
	}
}

func pngFileHeader(t *testing.T, filename string, width, height int) *multipart.FileHeader {
	t.Helper()

	var buffer bytes.Buffer
	if err := png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, width, height))); err != nil {
		t.Fatalf("encode PNG fixture: %v", err)
	}
	return testFileHeader(t, filename, buffer.Bytes())
}
