package gameforge

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
)

// PowChallenge is a single hashcash-style proof-of-work challenge.
// The nonce to find must satisfy: hex(sha256(salt || dec(nonce))) starts with target.
type PowChallenge struct {
	Salt   string `json:"salt"`
	Target string `json:"target"`
}

// PowFile is the container served by gameforge (OGame v13 login).
// The "instrumentation" field is an orthogonal browser fingerprinting payload
// that ogamed (headless) cannot execute; only the proof-of-work is needed.
type PowFile struct {
	Pow struct {
		Algorithm  string         `json:"algorithm"`
		Challenges []PowChallenge `json:"challenges"`
	} `json:"pow"`
	Instrumentation string `json:"instrumentation"`
}

// SolvePow finds the minimal nonce such that the sha256 digest of salt+dec(nonce)
// starts with the target hex prefix.
func SolvePow(salt, target string) int64 {
	var nonce int64
	for {
		h := sha256.Sum256([]byte(salt + strconv.FormatInt(nonce, 10)))
		if strings.HasPrefix(hex.EncodeToString(h[:]), target) {
			return nonce
		}
		nonce++
	}
}

// SolvePowParallel solves a single challenge by splitting the search space into
// contiguous windows processed by `workers` goroutines. It returns the minimal
// valid nonce, preserving equivalence with the sequential solver.
func SolvePowParallel(salt, target string, workers int) int64 {
	if workers < 1 {
		workers = 1
	}
	window := int64(2_000_000)
	var start int64
	for {
		found := make(chan int64, workers)
		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func(offset int64) {
				defer wg.Done()
				for n := start + offset; n < start+window; n += int64(workers) {
					h := sha256.Sum256([]byte(salt + strconv.FormatInt(n, 10)))
					if strings.HasPrefix(hex.EncodeToString(h[:]), target) {
						found <- n
						return
					}
				}
			}(int64(w))
		}
		wg.Wait()
		close(found)

		best := int64(-1)
		for n := range found {
			if best == -1 || n < best {
				best = n
			}
		}
		if best != -1 {
			return best
		}
		start += window
	}
}

// ParsePowFile parses the .pow JSON container.
func ParsePowFile(by []byte) (*PowFile, error) {
	var pf PowFile
	if err := json.Unmarshal(by, &pf); err != nil {
		return nil, err
	}
	if pf.Pow.Algorithm != "sha-256" {
		return nil, errors.New("unsupported pow algorithm: " + pf.Pow.Algorithm)
	}
	if len(pf.Pow.Challenges) == 0 {
		return nil, errors.New("pow file has no challenges")
	}
	return &pf, nil
}

// SolvePowFile solves all challenges of a .pow container and returns the
// nonces in the same order.
func SolvePowFile(by []byte, workers int) ([]int64, error) {
	pf, err := ParsePowFile(by)
	if err != nil {
		return nil, err
	}
	solutions := make([]int64, len(pf.Pow.Challenges))
	for i, c := range pf.Pow.Challenges {
		if workers > 1 {
			solutions[i] = SolvePowParallel(c.Salt, c.Target, workers)
		} else {
			solutions[i] = SolvePow(c.Salt, c.Target)
		}
	}
	return solutions, nil
}
