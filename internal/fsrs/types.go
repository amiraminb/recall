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
	RequestRetention float64   // Target retention rate (default 0.88)
	MaximumInterval  int       // Maximum days between reviews (default 1825)
	W                []float64 // Model weights
}

// DefaultParameters returns the default FSRS-5 parameters
func DefaultParameters() Parameters {
	return Parameters{
		RequestRetention: 0.88,
		MaximumInterval:  1825,
		W: []float64{
			0.4072, 1.1829, 3.1262, 15.4722, // Initial stability for each rating
			7.2102, // Initial difficulty base
			0.5316, // Initial difficulty slope
			1.0651, // Difficulty delta after rating
			0.0234, // Difficulty mean reversion
			1.616,  // Recall stability growth base
			0.1544, // Stability damping exponent
			1.0824, // Recall retrievability sensitivity
			1.9813, // Forget stability base
			0.0953, // Forget difficulty exponent
			0.2975, // Forget stability exponent
			2.2261, // Forget retrievability sensitivity
			0.2553, // Hard penalty
			0.6368, // Easy bonus
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
