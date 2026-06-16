// internal/score/score.go
package score

import (
	"k8s.io/klog/v2"
)

// MaxScore is the upper bound of the normalized score.
const MaxScore = 100

// MinScore is the lower bound of the normalized score.
const MinScore = -100

// ScoreNormalizer holds the minimum and maximum values for normalization,
// provides a normalize library to generate scores for AddOnPlacementScore.
type ScoreNormalizer struct {
	min float64
	max float64
}

// NewScoreNormalizer creates a new instance of ScoreNormalizer with given min and max values.
func NewScoreNormalizer(min, max float64) *ScoreNormalizer {
	return &ScoreNormalizer{
		min: min,
		max: max,
	}
}

// Normalize normalizes a given value to the range -100 to 100 based on the min and max values.
func (s *ScoreNormalizer) Normalize(value float64) (score int32, err error) {
	if value > s.max {
		score = MaxScore
	} else if value <= s.min {
		score = MinScore
	} else {
		score = (int32)((MaxScore-MinScore)*(value-s.min)/(s.max-s.min) + MinScore)
	}
	klog.Infof("value = %v, min = %v, max = %v, score = %v", value, s.min, s.max, score)
	return score, nil
}