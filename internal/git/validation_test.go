package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name      string
		branch    string
		expectErr bool
	}{
		{name: "valid simple", branch: "feature-123", expectErr: false},
		{name: "valid with slash", branch: "feature/awesome", expectErr: false},
		{name: "valid with dots", branch: "release-1.2.3", expectErr: false},
		{name: "valid underscore", branch: "my_feature", expectErr: false},
		{name: "empty", branch: "", expectErr: true},
		{name: "with space", branch: "feature 123", expectErr: true},
		{name: "with tilde", branch: "feature~123", expectErr: true},
		{name: "with caret", branch: "feature^123", expectErr: true},
		{name: "with colon", branch: "feature:123", expectErr: true},
		{name: "with question", branch: "feature?123", expectErr: true},
		{name: "with asterisk", branch: "feature*123", expectErr: true},
		{name: "with bracket", branch: "feature[123", expectErr: true},
		{name: "with backslash", branch: "feature\\123", expectErr: true},
		{name: "starts with dash", branch: "-feature", expectErr: true},
		{name: "ends with slash", branch: "feature/", expectErr: true},
		{name: "starts with slash", branch: "/feature", expectErr: true},
		{name: "ends with .lock", branch: "feature.lock", expectErr: true},
		{name: "consecutive slashes", branch: "feature//123", expectErr: true},
		{name: "just dot", branch: ".", expectErr: true},
		{name: "just double dot", branch: "..", expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBranchName(tt.branch)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
