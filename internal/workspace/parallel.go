package workspace

import (
	"sync"
)

// ParallelResult represents the result of a parallel operation
type ParallelResult struct {
	RepoName string
	Error    error
}

// ExecuteParallel runs a function for each repo in parallel
func ExecuteParallel(repos []string, fn func(repo string) error) []ParallelResult {
	var wg sync.WaitGroup
	results := make([]ParallelResult, len(repos))

	for i, repo := range repos {
		wg.Add(1)
		go func(index int, repoName string) {
			defer wg.Done()
			err := fn(repoName)
			results[index] = ParallelResult{
				RepoName: repoName,
				Error:    err,
			}
		}(i, repo)
	}

	wg.Wait()
	return results
}
