package janitor

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestUploadExpirer(t *testing.T) {
	d1 := time.Hour * 24           // 1 day
	d2 := time.Hour * 24 * 90      // 3 months
	d3 := time.Hour * 24 * 180     // 6 months
	d4 := time.Hour * 24 * 365 * 2 // 2 years

	globalPolicies := []dbstore.ConfigurationPolicy{
		{
			Type:              "GIT_TREE",
			Pattern:           "*",
			RetentionEnabled:  true,
			RetentionDuration: &d2,
		},
		{
			Type:              "GIT_TAG",
			Pattern:           "*",
			RetentionEnabled:  true,
			RetentionDuration: &d3,
		},
		{
			Type:              "GIT_TREE",
			Pattern:           "main",
			RetentionEnabled:  true,
			RetentionDuration: nil, // indefinite
		},
	}

	repositoryPolicies := map[int][]dbstore.ConfigurationPolicy{
		50: {
			dbstore.ConfigurationPolicy{
				Type:                      "GIT_TREE",
				Pattern:                   "ef/*",
				RetentionEnabled:          true,
				RetainIntermediateCommits: true,
				RetentionDuration:         &d4,
			},
		},
		51: {},
		52: {},
		53: {
			dbstore.ConfigurationPolicy{
				Type:              "GIT_COMMIT",
				Pattern:           "*",
				RetentionEnabled:  true,
				RetentionDuration: &d1,
			},
		},
	}

	now := timeutil.Now()
	t1 := now.Add(-time.Hour)                 // 1 hour old
	t2 := now.Add(-time.Hour * 24 * 7)        // 1 week ago
	t3 := now.Add(-time.Hour * 24 * 30 * 5)   // 5 months ago
	t4 := now.Add(-time.Hour * 24 * 30 * 9)   // 9 months ago
	t5 := now.Add(-time.Hour * 24 * 30 * 18)  // 18 months ago
	t6 := now.Add(-time.Hour * 24 * 365 * 2)  // 3 years ago
	t8 := now.Add(-time.Hour * 24 * 365 * 15) // 15 years ago

	branchMap := map[string]map[string]string{
		"deadbeef01": {"develop": "deadbeef01"},
		"deadbeef02": {"develop": "deadbeef01", "feat/blank": "deadbeef02"},
		"deadbeef03": {"develop": "deadbeef01"},
		"deadbeef04": {"develop": "deadbeef01"},
		"deadbeef05": {"develop": "deadbeef01"},
		"deadbeef06": {"es/feature-z": "deadbeef06"},
		"deadbeef07": {"ef/feature-x": "deadbeef07"},
		"deadbeef08": {"ef/feature-x": "deadbeef07"},
		"deadbeef09": {"ef/feature-y": "deadbeef09"},
		"deadbeef10": {"ef/feature-w": "deadbeef10"},
		"deadbeef11": {"main": "deadbeef11"},
		"deadbeef12": {"main": "deadbeef11"},
	}

	tagMap := map[string][]string{
		"deadbeef01": nil,
		"deadbeef02": nil,
		"deadbeef03": nil,
		"deadbeef04": {"v1.2.3"},
		"deadbeef05": {"v1.2.2"},
		"deadbeef06": nil,
		"deadbeef07": nil,
		"deadbeef08": nil,
		"deadbeef09": nil,
		"deadbeef10": nil,
		"deadbeef11": nil,
		"deadbeef12": nil,
	}

	uploads := []dbstore.Upload{
		//
		// Repository 50
		//
		// Commit graph sketch:
		//
		//    05 ------ 04 ------ 03 ------ 02 ------ 01
		//     \         \                   \         \
		//      v1.2.2    v1.2.3              \         develop
		//                                     feat/blank
		//
		//              08 ---- 07
		//  09                   \                     06
		//   \                   ef/feature-x           \
		//    ef/feature-y                              es/feature-z

		// 1 week old
		// tip of develop (PROTECTED, younger than 3 months)
		{ID: 1, RepositoryID: 50, Commit: "deadbeef01", State: "completed", FinishedAt: &t2},

		// 1 week old
		// on develop (UNPROTECTED, not tip)
		// tip of feat/blank (PROTECTED, younger than 3 months)
		{ID: 2, RepositoryID: 50, Commit: "deadbeef02", State: "completed", FinishedAt: &t2},

		// 5 months old
		// on develop (UNPROTECTED, not tip)
		{ID: 3, RepositoryID: 50, Commit: "deadbeef03", State: "completed", FinishedAt: &t3},

		// 5 months old
		// on develop (UNPROTECTED, not tip)
		// tag v1.2.3 (PROTECTED, younger than 6 months)
		{ID: 4, RepositoryID: 50, Commit: "deadbeef04", State: "completed", FinishedAt: &t3},

		// 9 months old
		// on develop (UNPROTECTED, not tip)
		// tag v1.2.2 (UNPROTECTED, older than 6 months)
		{ID: 5, RepositoryID: 50, Commit: "deadbeef05", State: "completed", FinishedAt: &t4},

		// 5 months old
		// tip of es/feature-z (UNPROTECTED, older than 3 months)
		{ID: 6, RepositoryID: 50, Commit: "deadbeef06", State: "completed", FinishedAt: &t3},

		// 9 months old
		// tip of ef/feature-x (PROTECTED, younger than 2 years)
		{ID: 7, RepositoryID: 50, Commit: "deadbeef07", State: "completed", FinishedAt: &t4},

		// 18 months old
		// on ef/feature-x (PROTECTED, younger than 2 years)
		{ID: 8, RepositoryID: 50, Commit: "deadbeef08", State: "completed", FinishedAt: &t5},

		// 3 years old
		// tip of ef/feature-y (UNPROTECTED, older than 2 years)
		{ID: 9, RepositoryID: 50, Commit: "deadbeef09", State: "completed", FinishedAt: &t6},

		//
		// Repository 51

		// 9 months old
		// tip of ef/feature-w (UNPROTECTED, policy does not apply to this repo)
		{ID: 10, RepositoryID: 51, Commit: "deadbeef10", State: "completed", FinishedAt: &t4},

		//
		// Repository 52

		// 15 years old
		// tip of main (PROTECTED, no duration)
		{ID: 11, RepositoryID: 52, Commit: "deadbeef11", State: "completed", FinishedAt: &t8},

		// 15 years old
		// on main (UNPROTECTED, not tip)
		{ID: 12, RepositoryID: 52, Commit: "deadbeef12", State: "completed", FinishedAt: &t8},

		//
		// Repository 53

		// 1 hour old
		// covered by catch-all (PROTECTED, younger than 1 day)
		{ID: 13, RepositoryID: 53, Commit: "deadbeef13", State: "completed", FinishedAt: &t1},
	}

	dbStore := testUploadExpirerMockStore(globalPolicies, repositoryPolicies, uploads)
	gitserverClient := testUploadExpirerMockGitserverClient(branchMap, tagMap)

	uploadExpirer := &uploadExpirer{
		dbStore:                dbStore,
		gitserverClient:        gitserverClient,
		metrics:                nil,
		repositoryProcessDelay: 24 * time.Hour,
		repositoryBatchSize:    100,
		uploadProcessDelay:     24 * time.Hour,
		uploadBatchSize:        100,
	}

	if err := uploadExpirer.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error from handle: %s", err)
	}

	var protectedIDs []int
	for _, call := range dbStore.UpdateUploadRetentionFunc.History() {
		protectedIDs = append(protectedIDs, call.Arg1...)
	}
	sort.Ints(protectedIDs)

	var expiredIDs []int
	for _, call := range dbStore.UpdateUploadRetentionFunc.History() {
		expiredIDs = append(expiredIDs, call.Arg2...)
	}
	sort.Ints(expiredIDs)

	if diff := cmp.Diff([]int{1, 2, 4, 7, 8, 11, 13}, protectedIDs); diff != "" {
		t.Errorf("unexpected protected upload identifiers (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]int{3, 5, 6, 9, 10, 12}, expiredIDs); diff != "" {
		t.Errorf("unexpected expired upload identifiers (-want +got):\n%s", diff)
	}
}

func testUploadExpirerMockStore(
	globalPolicies []dbstore.ConfigurationPolicy,
	repositoryPolicies map[int][]dbstore.ConfigurationPolicy,
	uploads []dbstore.Upload,
) *MockDBStore {
	repositoryIDs := make([]int, 0, len(repositoryPolicies))
	for repositoryID := range repositoryPolicies {
		repositoryIDs = append(repositoryIDs, repositoryID)
	}

	state := &uploadExpirerMockStore{
		uploads:            uploads,
		repositoryIDs:      repositoryIDs,
		globalPolicies:     globalPolicies,
		repositoryPolicies: repositoryPolicies,
		protected:          map[int]time.Time{},
		expired:            map[int]struct{}{},
	}

	dbStore := NewMockDBStore()
	dbStore.RepositoryIDsForRetentionScanFunc.SetDefaultHook(state.RepositoryIDsForRetentionScan)
	dbStore.GetConfigurationPoliciesFunc.SetDefaultHook(state.GetConfigurationPolicies)
	dbStore.GetUploadsFunc.SetDefaultHook(state.GetUploads)
	dbStore.CommitsVisibleToUploadFunc.SetDefaultHook(state.CommitsVisibleToUpload)
	dbStore.UpdateUploadRetentionFunc.SetDefaultHook(state.UpdateUploadRetention)
	return dbStore
}

type uploadExpirerMockStore struct {
	uploads            []dbstore.Upload
	repositoryIDs      []int
	globalPolicies     []dbstore.ConfigurationPolicy
	repositoryPolicies map[int][]dbstore.ConfigurationPolicy
	protected          map[int]time.Time
	expired            map[int]struct{}
}

func (s *uploadExpirerMockStore) GetUploads(ctx context.Context, opts dbstore.GetUploadsOptions) ([]dbstore.Upload, int, error) {
	var filtered []dbstore.Upload
	for _, upload := range s.uploads {
		if upload.RepositoryID != opts.RepositoryID {
			continue
		}
		if _, ok := s.expired[upload.ID]; ok {
			continue
		}
		if lastScanned, ok := s.protected[upload.ID]; ok && !lastScanned.Before(*opts.LastRetentionScanBefore) {
			continue
		}

		filtered = append(filtered, upload)
	}

	if len(filtered) > opts.Limit {
		filtered = filtered[:opts.Limit]
	}

	return filtered, len(s.uploads), nil
}

func (s *uploadExpirerMockStore) CommitsVisibleToUpload(ctx context.Context, uploadID int) ([]string, error) {
	for _, upload := range s.uploads {
		if upload.ID == uploadID {
			return []string{upload.Commit}, nil
		}
	}

	return nil, nil
}

func (s *uploadExpirerMockStore) UpdateUploadRetention(ctx context.Context, protectedIDs, expiredIDs []int) error {
	for _, id := range protectedIDs {
		s.protected[id] = time.Now()
	}

	for _, id := range expiredIDs {
		s.expired[id] = struct{}{}
	}

	return nil
}

func (state *uploadExpirerMockStore) RepositoryIDsForRetentionScan(ctx context.Context, processDelay time.Duration, limit int) (scannedIDs []int, _ error) {
	if len(state.repositoryIDs) <= limit {
		scannedIDs, state.repositoryIDs = state.repositoryIDs, nil
	} else {
		scannedIDs, state.repositoryIDs = state.repositoryIDs[:limit], state.repositoryIDs[limit:]
	}

	return scannedIDs, nil
}

func (state *uploadExpirerMockStore) GetConfigurationPolicies(ctx context.Context, opts dbstore.GetConfigurationPoliciesOptions) ([]dbstore.ConfigurationPolicy, error) {
	if opts.RepositoryID == 0 {
		return state.globalPolicies, nil
	}

	policies, ok := state.repositoryPolicies[opts.RepositoryID]
	if !ok {
		return nil, errors.Errorf("unexpected repository argument %d", opts.RepositoryID)
	}

	return policies, nil
}

func testUploadExpirerMockGitserverClient(branchMap map[string]map[string]string, tagMap map[string][]string) *MockGitserverClient {
	gitserverClient := NewMockGitserverClient()

	gitserverClient.RefDescriptionsFunc.SetDefaultHook(func(ctx context.Context, repositoryID int) (map[string][]gitserver.RefDescription, error) {
		refDescriptions := map[string][]gitserver.RefDescription{}
		for commit, branches := range branchMap {
			for branch, tip := range branches {
				if tip != commit {
					continue
				}

				refDescriptions[commit] = append(refDescriptions[commit], gitserver.RefDescription{
					Name: branch,
					Type: gitserver.RefTypeBranch,
				})
			}
		}

		for commit, tags := range tagMap {
			for _, tag := range tags {
				refDescriptions[commit] = append(refDescriptions[commit], gitserver.RefDescription{
					Name: tag,
					Type: gitserver.RefTypeTag,
				})
			}
		}

		return refDescriptions, nil
	})

	gitserverClient.BranchesContainingFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, commit string) ([]string, error) {
		var branches []string
		for branch := range branchMap[commit] {
			branches = append(branches, branch)
		}

		return branches, nil
	})

	return gitserverClient
}
