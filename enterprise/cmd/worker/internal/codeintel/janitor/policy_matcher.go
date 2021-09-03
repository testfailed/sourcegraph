package janitor

import (
	"github.com/gobwas/glob"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
)

type matcherFunc func(policies []dbstore.ConfigurationPolicy) bool

// newTipPolicyMatcher returns a matcher function that tests a set of policies against the
// given commit, which defines the tips of the tags and branches in the given set of reference
// descriptions.
//
// The resulting matcher does not check policies for enabled data retention, nor does it check
// if it covers the time range containing the commit. Such policies must be removed from the
// working set prior to invocation of this function.
func newTipPolicyMatcher(patterns map[string]glob.Glob, commit string, refDescriptions []gitserver.RefDescription) matcherFunc {
	branches, tags := refNamesByType(refDescriptions)

	return func(policies []dbstore.ConfigurationPolicy) bool {
		for _, policy := range policies {
			if policy.Type == "GIT_COMMIT" && patterns[policy.Pattern].Match(commit) {
				return true
			} else if policy.Type == "GIT_TAG" && patternMatchesAnyValue(patterns[policy.Pattern], tags) {
				return true
			} else if policy.Type == "GIT_TREE" && patternMatchesAnyValue(patterns[policy.Pattern], branches) {
				return true
			}
		}

		return false
	}
}

// newContainsPolicyMatcher returns a matcher function that tests a set of policies against the
// branches that contain the given commit.
//
// The resulting matcher does not check policies for enabled data retention, nor does it check if
// it covers the time range containing the commit, nor does it check if the intermediate commits
// of a branch are covered by the policy. Such policies must be removed from the working set prior
// to invocation of this function.
func newContainsPolicyMatcher(patterns map[string]glob.Glob, commit string, branches []string) matcherFunc {
	return func(policies []dbstore.ConfigurationPolicy) bool {
		for _, policy := range policies {
			if patternMatchesAnyValue(patterns[policy.Pattern], branches) {
				return true
			}
		}

		return false
	}
}

// refNamesByType returns slices of names of the given references description bucketed by their type.
func refNamesByType(refDescriptions []gitserver.RefDescription) (branches, tags []string) {
	branches = make([]string, 0, len(refDescriptions))
	tags = make([]string, 0, len(refDescriptions))

	for _, refDescription := range refDescriptions {
		if refDescription.Type == gitserver.RefTypeBranch {
			branches = append(branches, refDescription.Name)
		} else if refDescription.Type == gitserver.RefTypeTag {
			tags = append(tags, refDescription.Name)
		}
	}

	return branches, tags
}

// patternMatchesAnyValue returns true if the given pattern matches at least one of the given values.
func patternMatchesAnyValue(pattern glob.Glob, values []string) bool {
	for _, value := range values {
		if pattern.Match(value) {
			return true
		}
	}

	return false
}
