package archiver

import "testing"

func Test_validateArchiveSelection(t *testing.T) {
	type args struct {
		extensions []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"zip", args{[]string{"zip"}}, false},
		{"xyz", args{[]string{"zip", "xyz"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateArchiveSelection(tt.args.extensions); (err != nil) != tt.wantErr {
				t.Errorf("validateArchiveSelection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateExtension(t *testing.T) {
	type args struct {
		extension string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"zip", args{"zip"}, false},
		{"tar", args{"tar"}, false},
		{"error", args{"ar"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateExtension(tt.args.extension); (err != nil) != tt.wantErr {
				t.Errorf("validateExtension() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
