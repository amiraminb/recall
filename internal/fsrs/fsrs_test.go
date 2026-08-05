package fsrs

import (
	"math"
	"testing"
	"time"
)

var epoch = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

// matureCard returns a card that has been read and reviewed Good several times,
// each review landing exactly on the due date.
func matureCard(t *testing.T, f *FSRS, reviews int) Card {
	t.Helper()

	card := f.Review(NewCard(), Good, epoch)
	for range reviews {
		card = f.Review(card, Good, card.Due)
	}
	return card
}

func TestIntervalIncreasesWithRating(t *testing.T) {
	f := NewScheduler()
	card := matureCard(t, f, 3)
	at := card.Due

	var prev int
	for _, rating := range []Rating{Again, Hard, Good, Easy} {
		got := f.Review(card, rating, at)
		days := int(got.Due.Sub(at).Hours() / 24)

		if days < prev {
			t.Errorf("rating %d gave interval %d days, shorter than the previous rating's %d",
				rating, days, prev)
		}
		prev = days
	}
}

func TestEasyBeatsGood(t *testing.T) {
	f := NewScheduler()
	card := matureCard(t, f, 3)
	at := card.Due

	good := f.Review(card, Good, at)
	easy := f.Review(card, Easy, at)

	if easy.Stability <= good.Stability {
		t.Errorf("Easy stability %.2f must exceed Good %.2f; w[16] is likely below 1",
			easy.Stability, good.Stability)
	}
	if !easy.Due.After(good.Due) {
		t.Errorf("Easy due %s must be later than Good due %s",
			easy.Due.Format(time.DateOnly), good.Due.Format(time.DateOnly))
	}
}

func TestAgainShortensAndLapses(t *testing.T) {
	f := NewScheduler()
	card := matureCard(t, f, 3)
	at := card.Due

	got := f.Review(card, Again, at)

	if got.Stability >= card.Stability {
		t.Errorf("Again must reduce stability: got %.2f, was %.2f", got.Stability, card.Stability)
	}
	if got.Lapses != card.Lapses+1 {
		t.Errorf("Again must increment lapses: got %d, want %d", got.Lapses, card.Lapses+1)
	}
	if got.State != Relearn {
		t.Errorf("Again must move card to Relearn, got state %d", got.State)
	}
}

// The interval formula inverts the forgetting curve, so R(S, S) must equal the
// 0.9 anchor both use. An exponential curve here silently breaks that identity.
func TestRetrievabilityMatchesIntervalFormula(t *testing.T) {
	f := NewScheduler()

	for _, stability := range []float64{1, 10, 50, 200} {
		got := f.retrievability(stability, stability)
		if math.Abs(got-0.9) > 1e-9 {
			t.Errorf("retrievability(S=%.0f, t=S) = %.6f, want 0.9", stability, got)
		}
	}

	f.Params.RequestRetention = 0.9
	for _, stability := range []float64{10, 50, 200} {
		days := f.nextInterval(stability)
		if math.Abs(float64(days)-stability) > 1 {
			t.Errorf("at 0.9 retention, interval for S=%.0f should be ~%.0f days, got %d",
				stability, stability, days)
		}
	}
}

func TestRetrievabilityDecaysMonotonically(t *testing.T) {
	f := NewScheduler()

	prev := 1.0
	for _, elapsed := range []float64{1, 5, 10, 30, 90, 365, 3650} {
		got := f.retrievability(20, elapsed)
		if got > prev {
			t.Errorf("retrievability rose from %.4f to %.4f at %.0f days", prev, got, elapsed)
		}
		if got < 0 || got > 1 {
			t.Errorf("retrievability %.4f out of range at %.0f days", got, elapsed)
		}
		prev = got
	}
}

// A review far past its due date means more forgetting, so stability must grow
// more than an on-time review of the same card at the same rating.
func TestLateReviewEarnsMoreStability(t *testing.T) {
	f := NewScheduler()
	card := matureCard(t, f, 2)

	onTime := f.Review(card, Good, card.Due)
	late := f.Review(card, Good, card.Due.AddDate(0, 0, 90))

	if late.Stability <= onTime.Stability {
		t.Errorf("late review stability %.2f should exceed on-time %.2f",
			late.Stability, onTime.Stability)
	}
}

func TestVeryLateReviewStaysBounded(t *testing.T) {
	f := NewScheduler()
	card := matureCard(t, f, 2)

	got := f.Review(card, Good, card.Due.AddDate(5, 0, 0))
	days := int(got.Due.Sub(card.Due.AddDate(5, 0, 0)).Hours() / 24)

	if days > f.Params.MaximumInterval {
		t.Errorf("interval %d days exceeds maximum %d", days, f.Params.MaximumInterval)
	}
	if math.IsNaN(got.Stability) || math.IsInf(got.Stability, 0) {
		t.Errorf("stability became %v after a 5-year-late review", got.Stability)
	}
}

// Reviewing before the due date must not be punished with a shorter interval
// than the card already had.
func TestEarlyReviewDoesNotRegress(t *testing.T) {
	f := NewScheduler()
	card := matureCard(t, f, 3)

	early := card.LastReview.AddDate(0, 0, 1)
	got := f.Review(card, Good, early)

	if got.Stability < card.Stability {
		t.Errorf("early Good review dropped stability from %.2f to %.2f",
			card.Stability, got.Stability)
	}
}

func TestDifficultyStaysInRange(t *testing.T) {
	f := NewScheduler()

	card := f.Review(NewCard(), Again, epoch)
	for range 50 {
		card = f.Review(card, Again, card.Due)
		if card.Difficulty < 1 || card.Difficulty > 10 {
			t.Fatalf("difficulty %.4f left [1,10] after repeated Again", card.Difficulty)
		}
	}

	if card.Difficulty <= 9 {
		t.Errorf("difficulty should approach 10 after 50 lapses, got %.2f", card.Difficulty)
	}

	for range 50 {
		card = f.Review(card, Easy, card.Due)
		if card.Difficulty < 1 || card.Difficulty > 10 {
			t.Fatalf("difficulty %.4f left [1,10] after repeated Easy", card.Difficulty)
		}
	}
}

func TestStabilityNeverNaN(t *testing.T) {
	f := NewScheduler()
	ratings := []Rating{Good, Again, Hard, Easy, Again, Again, Good, Easy, Hard}

	card := NewCard()
	now := epoch
	for i, rating := range ratings {
		card = f.Review(card, rating, now)

		if math.IsNaN(card.Stability) || math.IsInf(card.Stability, 0) {
			t.Fatalf("stability %v after review %d (rating %d)", card.Stability, i, rating)
		}
		if card.Stability <= 0 {
			t.Fatalf("stability %.4f not positive after review %d", card.Stability, i)
		}
		if math.IsNaN(card.Difficulty) {
			t.Fatalf("difficulty NaN after review %d", i)
		}
		now = card.Due
	}
}

func TestZeroStabilityCardRecovers(t *testing.T) {
	f := NewScheduler()

	corrupt := Card{State: Review, Stability: 0, Difficulty: 5, LastReview: epoch}
	got := f.Review(corrupt, Good, epoch.AddDate(0, 0, 10))

	if got.Stability <= 0 || math.IsNaN(got.Stability) {
		t.Errorf("zero-stability card did not recover: got %v", got.Stability)
	}
}

func TestFirstReadIntervals(t *testing.T) {
	f := NewScheduler()

	want := map[Rating]int{Again: 1, Hard: 1, Good: 3, Easy: 15}
	for rating, wantDays := range want {
		card := f.Review(NewCard(), rating, epoch)
		days := int(card.Due.Sub(epoch).Hours() / 24)

		if days != wantDays {
			t.Errorf("first read rated %d scheduled %d days out, want %d", rating, days, wantDays)
		}
		if card.Reps != 1 {
			t.Errorf("first read should set Reps=1, got %d", card.Reps)
		}
	}
}

func TestNewCardAgainEntersLearning(t *testing.T) {
	f := NewScheduler()

	card := f.Review(NewCard(), Again, epoch)
	if card.State != Learning {
		t.Errorf("first read rated Again should enter Learning, got state %d", card.State)
	}
}

func TestInvalidRatingLeavesCardUntouched(t *testing.T) {
	f := NewScheduler()
	card := matureCard(t, f, 2)

	for _, rating := range []Rating{Rating(0), Rating(5), Rating(-1)} {
		got := f.Review(card, rating, card.Due)
		if got != card {
			t.Errorf("rating %d mutated the card", rating)
		}
	}
}

func TestIntervalRespectsMaximum(t *testing.T) {
	f := NewScheduler()
	f.Params.MaximumInterval = 30

	card := NewCard()
	now := epoch
	for range 12 {
		card = f.Review(card, Easy, now)
		days := int(card.Due.Sub(now).Hours() / 24)

		if days > 30 {
			t.Fatalf("interval %d days exceeds configured maximum of 30", days)
		}
		now = card.Due
	}
}

func TestLowerRetentionMeansLongerIntervals(t *testing.T) {
	loose, tight := NewScheduler(), NewScheduler()
	loose.Params.RequestRetention = 0.8
	tight.Params.RequestRetention = 0.95

	if loose.nextInterval(50) <= tight.nextInterval(50) {
		t.Errorf("0.8 retention gave %d days, not longer than 0.95's %d",
			loose.nextInterval(50), tight.nextInterval(50))
	}
}

func TestElapsedDaysIgnoresClockSkew(t *testing.T) {
	f := NewScheduler()
	card := Card{State: Review, Stability: 10, Difficulty: 5, LastReview: epoch}

	if got := f.elapsedDays(card, epoch.AddDate(0, 0, -5)); got != 0 {
		t.Errorf("a review timestamped before the last review should give 0 elapsed days, got %v", got)
	}
	if got := f.elapsedDays(NewCard(), epoch); got != 0 {
		t.Errorf("an unreviewed card should give 0 elapsed days, got %v", got)
	}
}
