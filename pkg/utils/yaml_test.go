package utils

import (
	"testing"
)

func TestSplitJSONRef(t *testing.T) {
	type args struct {
		refStr string
	}
	tests := []struct {
		name         string
		args         args
		wantLocation string
		wantJsonPath string
		wantErr      bool
	}{
		{
			"absolute file path",
			args{
				"/hello/world.yaml",
			},
			"/hello/world.yaml",
			"$",
			false,
		},
		{
			"relative file path",
			args{
				"../relative.yaml",
			},
			"../relative.yaml",
			"$",
			false,
		},
		{
			"file with fragment",
			args{
				"hello.yaml#/foo",
			},
			"hello.yaml",
			"$.foo",
			false,
		},
		{
			"no path no fragment",
			args{
				"",
			},
			"",
			"$",
			false,
		},
		{
			"deep fragment only",
			args{
				"#/foo/bar/fizz",
			},
			"",
			"$.foo.bar.fizz",
			false,
		},
		{
			"root fragment #",
			args{
				"#",
			},
			"",
			"$",
			false,
		},
		{
			"root fragment #/",
			args{
				"#/",
			},
			"",
			"$",
			false,
		},
		{
			"file with deep pointer",
			args{
				"hello.yaml#/foo/bar/fizz",
			},
			"hello.yaml",
			"$.foo.bar.fizz",
			false,
		},
		{
			"remote JSONRef",
			args{
				"https://example.com/hello.yaml#/foo/bar/fizz",
			},
			"",
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLocation, gotJsonPath, err := SplitJSONRef(tt.args.refStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitJSONRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotLocation != tt.wantLocation {
				t.Errorf("SplitJSONRef() gotLocation = %v, want %v", gotLocation, tt.wantLocation)
			}
			if gotJsonPath != tt.wantJsonPath {
				t.Errorf("SplitJSONRef() gotJsonPath = %v, want %v", gotJsonPath, tt.wantJsonPath)
			}
		})
	}
}
