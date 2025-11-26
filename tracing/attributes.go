package tracing

import (
	"go.opentelemetry.io/otel/attribute"
)

// Standard attribute keys for gitops operations.
// These provide consistent naming across the codebase.
const (
	// Git operation attributes
	AttrGitURL        = "git.url"
	AttrGitRef        = "git.ref"
	AttrGitBranch     = "git.branch"
	AttrGitHash       = "git.hash"
	AttrGitDirectory  = "git.directory"
	AttrGitPreRef     = "git.pre_ref"
	AttrGitPostRef    = "git.post_ref"

	// Entrypoint attributes
	AttrEntrypointType      = "entrypoint.type"
	AttrEntrypointDirectory = "entrypoint.directory"
	AttrEntrypointCount     = "entrypoint.count"

	// Resource attributes
	AttrResourceType    = "resource.type"
	AttrResourceCount   = "resource.count"
	AttrResourceName    = "resource.name"
	AttrDiffCount       = "diff.count"
	AttrDiffCreated     = "diff.created"
	AttrDiffUpdated     = "diff.updated"
	AttrDiffDeleted     = "diff.deleted"
	AttrDiffReplaced    = "diff.replaced"

	// Directory attributes
	AttrWorkingDir = "dir.working"
	AttrPreDir     = "dir.pre"
	AttrPostDir    = "dir.post"

	// Operation result attributes
	AttrErrorOccurred = "error.occurred"
	AttrErrorMessage  = "error.message"
)

// GitURL creates an attribute for git repository URL.
func GitURL(url string) attribute.KeyValue {
	return attribute.String(AttrGitURL, url)
}

// GitRef creates an attribute for git reference.
func GitRef(ref string) attribute.KeyValue {
	return attribute.String(AttrGitRef, ref)
}

// GitBranch creates an attribute for git branch.
func GitBranch(branch string) attribute.KeyValue {
	return attribute.String(AttrGitBranch, branch)
}

// GitHash creates an attribute for git commit hash.
func GitHash(hash string) attribute.KeyValue {
	return attribute.String(AttrGitHash, hash)
}

// GitDirectory creates an attribute for git directory.
func GitDirectory(dir string) attribute.KeyValue {
	return attribute.String(AttrGitDirectory, dir)
}

// GitPreRef creates an attribute for pre-change git reference.
func GitPreRef(ref string) attribute.KeyValue {
	return attribute.String(AttrGitPreRef, ref)
}

// GitPostRef creates an attribute for post-change git reference.
func GitPostRef(ref string) attribute.KeyValue {
	return attribute.String(AttrGitPostRef, ref)
}

// EntrypointType creates an attribute for entrypoint type.
func EntrypointType(epType string) attribute.KeyValue {
	return attribute.String(AttrEntrypointType, epType)
}

// EntrypointDirectory creates an attribute for entrypoint directory.
func EntrypointDirectory(dir string) attribute.KeyValue {
	return attribute.String(AttrEntrypointDirectory, dir)
}

// EntrypointCount creates an attribute for entrypoint count.
func EntrypointCount(count int) attribute.KeyValue {
	return attribute.Int(AttrEntrypointCount, count)
}

// ResourceType creates an attribute for resource type.
func ResourceType(resType string) attribute.KeyValue {
	return attribute.String(AttrResourceType, resType)
}

// ResourceCount creates an attribute for resource count.
func ResourceCount(count int) attribute.KeyValue {
	return attribute.Int(AttrResourceCount, count)
}

// ResourceName creates an attribute for resource name.
func ResourceName(name string) attribute.KeyValue {
	return attribute.String(AttrResourceName, name)
}

// DiffCount creates an attribute for total diff count.
func DiffCount(count int) attribute.KeyValue {
	return attribute.Int(AttrDiffCount, count)
}

// DiffCreated creates an attribute for created resource count.
func DiffCreated(count int) attribute.KeyValue {
	return attribute.Int(AttrDiffCreated, count)
}

// DiffUpdated creates an attribute for updated resource count.
func DiffUpdated(count int) attribute.KeyValue {
	return attribute.Int(AttrDiffUpdated, count)
}

// DiffDeleted creates an attribute for deleted resource count.
func DiffDeleted(count int) attribute.KeyValue {
	return attribute.Int(AttrDiffDeleted, count)
}

// DiffReplaced creates an attribute for replaced resource count.
func DiffReplaced(count int) attribute.KeyValue {
	return attribute.Int(AttrDiffReplaced, count)
}

// WorkingDir creates an attribute for working directory.
func WorkingDir(dir string) attribute.KeyValue {
	return attribute.String(AttrWorkingDir, dir)
}

// PreDir creates an attribute for pre-change directory.
func PreDir(dir string) attribute.KeyValue {
	return attribute.String(AttrPreDir, dir)
}

// PostDir creates an attribute for post-change directory.
func PostDir(dir string) attribute.KeyValue {
	return attribute.String(AttrPostDir, dir)
}

// ErrorOccurred creates an attribute indicating whether an error occurred.
func ErrorOccurred(occurred bool) attribute.KeyValue {
	return attribute.Bool(AttrErrorOccurred, occurred)
}

// ErrorMessage creates an attribute for error message.
func ErrorMessage(msg string) attribute.KeyValue {
	return attribute.String(AttrErrorMessage, msg)
}
