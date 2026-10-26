package auth

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type contractOperation struct {
	OperationID string                 `yaml:"operationId"`
	MinRole     string                 `yaml:"x-min-role"`
	Security    *[]map[string][]string `yaml:"security"`
}

type contractDocument struct {
	Paths map[string]map[string]contractOperation `yaml:"paths"`
}

func TestPolicyMatchesOpenAPIContract(t *testing.T) {
	raw, err := os.ReadFile("../../../api/openapi.yaml")
	if err != nil {
		t.Fatalf("read openapi contract: %v", err)
	}
	var document contractDocument
	if err := yaml.Unmarshal(raw, &document); err != nil {
		t.Fatalf("parse openapi contract: %v", err)
	}

	policy := NewPolicy()
	methods := map[string]bool{
		"get": true, "post": true, "put": true, "patch": true, "delete": true,
	}
	seen := 0
	for path, operations := range document.Paths {
		for method, operation := range operations {
			if !methods[method] {
				continue
			}
			seen++
			if operation.OperationID == "" {
				t.Fatalf("%s %s: missing operationId", method, path)
			}
			name := strings.ToUpper(operation.OperationID[:1]) + operation.OperationID[1:]
			public := operation.MinRole == "public" ||
				(operation.Security != nil && len(*operation.Security) == 0)
			if public {
				if !policy.IsPublic(name) {
					t.Errorf("%s %s: contract is public but policy does not allow it", method, path)
				}
				continue
			}
			if policy.IsPublic(name) {
				t.Errorf("%s %s: policy allows publicly but contract does not", method, path)
			}
			if operation.MinRole == "" {
				t.Errorf("%s %s: missing x-min-role", method, path)
				continue
			}
			minRole, ok := policy.MinRole(name)
			if !ok {
				t.Errorf("%s %s: operation missing from policy", method, path)
				continue
			}
			if string(minRole) != operation.MinRole {
				t.Errorf("%s %s: policy role %q, contract %q", method, path, minRole, operation.MinRole)
			}
		}
	}
	if seen == 0 {
		t.Fatal("no operations found in openapi contract")
	}
}
