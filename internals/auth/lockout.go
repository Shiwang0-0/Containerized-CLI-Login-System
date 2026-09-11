package auth

import "time"

const (
	FailThreshold   = 5
	WarnAfter       = 3               // start warning once this many failures hit
	baseLockoutDur  = 2 * time.Minute // setting it to 2 min just for fast testing (normally it should be 15-30 min)
	maxLockoutDur   = 24 * time.Hour
	resetLevelAfter = 24 * time.Hour // if last lockout was longer ago than this, level resets
	maxLockoutLevel = 9              // cap the level too to avoid any overflow due to integer range limit
)

// nextLockout computes the duration + new level for a fresh lockout,
func NextLockout(lastLockoutAt time.Time, currentLevel int) (time.Duration, int) {
	level := 0

	// the last lockout happened before 24hr, that means the level of the waiting should increase
	// otherwise reset the level to 0
	/*
		base duration = x min
		First lockout                -----> level 0 ----->   x*(2^0) => x min lock duration
		Second lockout within 24hr   -----> level 1 -----> 	 x*(2^1) => 2x min lock duration
		Third lockout within 24hr    -----> level 2 ----->   x*(2^2) => 4x min lock duration
		.
		.
		.
		Nth lockout within 24hr      -----> level N-1 -----> min(24 hr, x*(2^(N-1))) => 24 hr lock duration
		.
		24 hr passed ===> level reset to 0
	*/

	if !lastLockoutAt.IsZero() && time.Since(lastLockoutAt) < resetLevelAfter {
		level = currentLevel + 1
	}

	if level > maxLockoutLevel {
		level = maxLockoutLevel
	}

	dur := baseLockoutDur << level       // base * 2^level
	if dur > maxLockoutDur || dur <= 0 { // cap the maximum wait to maxLockoutDur (we dont want infinte waiting)
		dur = maxLockoutDur
	}
	return dur, level
}
