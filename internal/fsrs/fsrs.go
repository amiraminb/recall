package fsrs

import (
	"math"
	"time"
)

// FSRS is the main scheduler.
type FSRS struct {
	Params Parameters
}

func NewScheduler() *FSRS {
	return &FSRS{Params: DefaultParameters()}
}

func (f *FSRS) Review(card Card, rating Rating, now time.Time) Card {
	if !rating.Valid() {
		return card
	}

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

// Linear damping shrinks the delta near d=10 so hard cards cannot pin themselves there.
func (f *FSRS) nextDifficulty(d float64, rating Rating) float64 {
	w := f.Params.W

	deltaD := -w[6] * (float64(rating) - 3)
	damped := deltaD * (10 - d) / 9
	d = f.meanReversion(f.initDifficulty(Easy), d+damped)

	return clamp(d, 1, 10)
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

	decay := f.Params.Decay
	interval := stability / curveFactor(decay) *
		(math.Pow(f.Params.RequestRetention, 1/decay) - 1)
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

// Power law, not exponential: nextInterval inverts this exact curve.
func (f *FSRS) retrievability(stability, elapsedDays float64) float64 {
	if stability <= 0 || elapsedDays <= 0 {
		return 1
	}

	decay := f.Params.Decay
	r := math.Pow(1+curveFactor(decay)*elapsedDays/stability, decay)
	return clamp(r, 0, 1)
}
