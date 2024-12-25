package stats

import "fmt"

// Implements Namespace and related functions
type Namespace struct {
	BucketName string `json:"bucket_name"`
	ScopeName  string `json:"scope_name"`
}

func NewNamespace(bucketName, scopeName string) (Namespace, error) {
	namespace := Namespace{}

	if bucketName == "" {
		bucketName = "*"
	}
	if scopeName == "" {
		scopeName = "*"
	}
	switch bucketName {
	case "*":
		return namespace, fmt.Errorf("wildcard not allowed")

	default:
		namespace.BucketName = bucketName
	}

	switch scopeName {
	case "*":
		return namespace, fmt.Errorf("wildcard not allowed")
	default:
		namespace.ScopeName = scopeName
	}
	return namespace, nil
}

func (namespace Namespace) String() string {
	return fmt.Sprintf("%s/%s", namespace.BucketName, namespace.ScopeName)
}

func (namespace Namespace) IsWildcard() bool {
	return (namespace.ScopeName == "*")
}

func (n1 Namespace) ExactEquals(n2 Namespace) bool {
	return (n1.BucketName == n2.BucketName) && (n1.ScopeName == n2.ScopeName)
}

func (n1 Namespace) Match(n2 Namespace) bool {
	if n1.BucketName != n2.BucketName {
		return false
	}

	if n1.IsWildcard() || n2.IsWildcard() {
		return true
	}

	return (n1.ScopeName == n2.ScopeName)
}
