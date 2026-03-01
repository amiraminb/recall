package fsrs

import (
	"math"
	"time"
)

const (
	// DECAY/FACTOR are the standard FSRS forgetting-curve constants.
	decay  = -0.5
	factor = 19.0 / 81.0
)

// FSRS is the main scheduler.
type FSRS struct {
	Params Parameters
}

func NewScheduler() *FSRS {
	return &FSRS{Params: DefaultParameters()}
}

func (f *FSRS) Review(card Card, rating Rating, now time.Time) Card {
	elapsedDays := f.elapsedDays(card, now)
	retrievability := f.retrievability(card.Stability, elapsedDays)

	card.LastReview = now
	card.Reps++

	if card.State == New {
		card = f.initializeCard(card, rating)
	} else {
		card = f.updateCard(card, rating, retrievability)
	}

	card.Due = now.AddDate(0, 0, f.nextInterval(card.Stability))

	return card
}

func (f *FSRS) initializeCard(card Card, rating Rating) Card {
	card.Difficulty = f.initDifficulty(rating)
	card.Stability = f.initStability(rating)

	if rating == Again {
		card.State = Learning
	} else {
		card.State = Review
	}

	return card
}

func (f *FSRS) updateCard(card Card, rating Rating, retrievability float64) Card {
	if rating == Again {
		card.Lapses++
		card.State = Relearn
		card.Stability = f.nextForgetStability(card.Difficulty, card.Stability, retrievability)
	} else {
		card.State = Review
		card.Stability = f.nextRecallStability(card, rating, retrievability)
	}

	card.Difficulty = f.nextDifficulty(card.Difficulty, rating)

	return card
}

func (f *FSRS) initDifficulty(rating Rating) float64 {
	w := f.Params.W
	d := w[4] - math.Exp(w[5]*float64(rating-1)) + 1
	return clamp(d, 1, 10)
}

func (f *FSRS) initStability(rating Rating) float64 {
	return f.Params.W[rating-1]
}

func (f *FSRS) nextDifficulty(d float64, rating Rating) float64 {
	w := f.Params.W

	// Difficulty update and mean reversion (FSRS-5).
	d = d - w[6]*(float64(rating)-3)
	d = f.meanReversion(w[4], d)
	d = clamp(d, 1, 10)
	return d
}

func (f *FSRS) nextRecallStability(card Card, rating Rating, retrievability float64) float64 {
	w := f.Params.W
	s := card.Stability
	d := card.Difficulty

	// Guard against zero or negative stability which causes NaN
	if s <= 0 {
		s = f.initStability(rating)
	}

	hardPenalty := 1.0
	if rating == Hard {
		hardPenalty = w[15]
	}

	easyBonus := 1.0
	if rating == Easy {
		easyBonus = w[16]
	}

	newS := s * (1 + math.Exp(w[8])*
		(11-d)*
		math.Pow(s, -w[9])*
		(math.Exp(w[10]*(1-retrievability))-1)*
		hardPenalty*
		easyBonus)

	return newS
}

func (f *FSRS) nextForgetStability(d, s, retrievability float64) float64 {
	w := f.Params.W

	if s <= 0 {
		s = 1
	}

	newS := w[11] *
		math.Pow(d, -w[12]) *
		(math.Pow(s+1, w[13]) - 1) *
		math.Exp(w[14]*(1-retrievability))

	return clamp(newS, 0.1, s)
}

func (f *FSRS) nextInterval(stability float64) int {
	if stability <= 0 {
		return 1
	}

	interval := stability / factor * (math.Pow(f.Params.RequestRetention, 1/decay) - 1)
	days := int(math.Round(interval))
	days = min(max(days, 1), f.Params.MaximumInterval)

	return days
}

func (f *FSRS) meanReversion(init, current float64) float64 {
	w := f.Params.W
	return w[7]*init + (1-w[7])*current
}

func (f *FSRS) elapsedDays(card Card, now time.Time) float64 {
	if card.LastReview.IsZero() {
		return 0
	}

	days := now.Sub(card.LastReview).Hours() / 24
	if days < 0 {
		return 0
	}

	return days
}

func (f *FSRS) retrievability(stability, elapsedDays float64) float64 {
	if stability <= 0 {
		return 1
	}
	if elapsedDays <= 0 {
		return 1
	}

	r := math.Exp(math.Log(0.9) * elapsedDays / stability)
	return clamp(r, 0, 1)
}
