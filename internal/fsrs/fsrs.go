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
	card.LastReview = now
	card.Reps++

	if card.State == New {
		card = f.initializeCard(card, rating)
	} else {
		card = f.updateCard(card, rating)
	}

	interval := f.nextInterval(card.Stability)
	card.Due = now.AddDate(0, 0, interval)

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

func (f *FSRS) updateCard(card Card, rating Rating) Card {
	if rating == Again {
		card.Lapses++
		card.State = Relearn
		card.Stability = f.nextForgetStability(card.Difficulty, card.Stability)
	} else {
		card.State = Review
		card.Stability = f.nextRecallStability(card, rating)
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
	delta := d - w[4]*(math.Exp(w[5]*float64(rating-3))-1)
	d = clamp(delta, 1, 10)
	return d
}

func (f *FSRS) nextRecallStability(card Card, rating Rating) float64 {
	w := f.Params.W
	s := card.Stability
	d := card.Difficulty

	// Guard against zero or negative stability which causes NaN
	if s <= 0 {
		s = f.initStability(rating)
	}

	hardPenalty := 1.0
	if rating == Hard {
		hardPenalty = w[12]
	}

	easyBonus := 1.0
	if rating == Easy {
		easyBonus = w[13]
	}

	newS := s * (1 + math.Exp(w[8])*
		(11-d)*
		math.Pow(s, -w[9])*
		(math.Exp(w[10]*float64(1-f.Params.RequestRetention))-1)*
		hardPenalty*
		easyBonus)

	return newS
}

func (f *FSRS) nextForgetStability(d, s float64) float64 {
	w := f.Params.W
	return w[6] * math.Pow(d, -w[7]) * (math.Pow(s+1, w[8]) - 1)
}

func (f *FSRS) nextInterval(stability float64) int {
	interval := min(max(int(math.Round(stability*9*(1/f.Params.RequestRetention-1))), 1), f.Params.MaximumInterval)

	return interval
}
