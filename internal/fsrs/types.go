package fsrs

import "time"

// Rating represents how well the user recalled the topic
type Rating int

const (
	Again Rating = 1 // Complete failure to recall
	Hard  Rating = 2 // Recalled with difficulty
	Good  Rating = 3 // Recalled with some effort
	Easy  Rating = 4 // Recalled effortlessly
)

// State represents the learning state of a card
type State int

const (
	New      State = 0
	Learning State = 1
	Review   State = 2
	Relearn  State = 3
)

// Card holds the FSRS scheduling data for a topic
type Card struct {
	Due        time.Time `json:"due"`
	Stability  float64   `json:"stability"`
	Difficulty float64   `json:"difficulty"`
	Reps       int       `json:"reps"`
	Lapses     int       `json:"lapses"`
	State      State     `json:"state"`
	LastReview time.Time `json:"last_review"`
}

// Parameters holds FSRS algorithm parameters
type Parameters struct {
	RequestRetention float64   // Target retention rate (default 0.9)
	MaximumInterval  int       // Maximum days between reviews (default 365)
	W                []float64 // Model weights
}

// DefaultParameters returns the default FSRS-5 parameters
func DefaultParameters() Parameters {
	return Parameters{
		RequestRetention: 0.9,
		MaximumInterval:  365,
		W: []float64{
			0.4072, 1.1829, 3.1262, 15.4722, // Initial stability for each rating
			7.2102, // Difficulty base
			0.5316, // Difficulty multiplier
			1.0651, // Stability after failure
			0.0234, // Stability short-term
			1.616,  // Stability multiplier (success)
			0.1544, // Stability multiplier (failure)
			1.0824, // Difficulty-stability interaction
			1.9813, // Hard penalty
			0.0953, // Easy bonus
			0.2975, // Difficulty floor
			2.2261, // Difficulty ceiling
			0.2553, // Stability recovery
			0.6368, // Stability decay
		},
	}
}

// NewCard creates a new card with default values
func NewCard() Card {
	return Card{
		Due:        time.Now(),
		Stability:  0,
		Difficulty: 0,
		Reps:       0,
		Lapses:     0,
		State:      New,
		LastReview: time.Time{},
	}
}
