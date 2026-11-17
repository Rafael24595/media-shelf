package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"slices"
	"strings"
)

// actorTags are the tags that named actors before ADR-0013, with the name as credited. They
// move from tags to actors; tags of other people (directors, writers, composers) stay.
var actorTags = map[string]string{
	"adam-sandler": "Adam Sandler", "al-pacino": "Al Pacino", "arnold-schwarzenegger": "Arnold Schwarzenegger",
	"bill-murray": "Bill Murray", "bob-odenkirk": "Bob Odenkirk", "brad-pitt": "Brad Pitt",
	"bruce-willis": "Bruce Willis", "bud-spencer": "Bud Spencer", "cantinflas": "Cantinflas",
	"charlie-chaplin": "Charlie Chaplin", "charlie-sheen": "Charlie Sheen", "charlton-heston": "Charlton Heston",
	"christian-bale": "Christian Bale", "christopher-lee": "Christopher Lee", "clint-eastwood": "Clint Eastwood",
	"daniel-day-lewis": "Daniel Day-Lewis", "dean-martin": "Dean Martin", "denzel-washington": "Denzel Washington",
	"eddie-murphy": "Eddie Murphy", "jackie-chan": "Jackie Chan", "jerry-lewis": "Jerry Lewis", "jet-li": "Jet Li",
	"jim-carrey": "Jim Carrey", "john-wayne": "John Wayne", "julia-roberts": "Julia Roberts",
	"keanu-reeves": "Keanu Reeves", "kevin-costner": "Kevin Costner", "kurt-russell": "Kurt Russell",
	"leslie-nielsen": "Leslie Nielsen", "martin-lawrence": "Martin Lawrence", "mel-gibson": "Mel Gibson",
	"michael-douglas": "Michael Douglas", "nicolas-cage": "Nicolas Cage", "rob-schneider": "Rob Schneider",
	"robert-de-niro": "Robert De Niro", "robin-williams": "Robin Williams", "ryan-gosling": "Ryan Gosling",
	"sacha-baron-cohen": "Sacha Baron Cohen", "sean-connery": "Sean Connery", "stephen-chow": "Stephen Chow",
	"sylvester-stallone": "Sylvester Stallone", "terence-hill": "Terence Hill", "tom-cruise": "Tom Cruise",
	"tom-hanks": "Tom Hanks", "van-damme": "Jean-Claude Van Damme", "viggo-mortensen": "Viggo Mortensen",
	"vin-diesel": "Vin Diesel", "will-ferrell": "Will Ferrell",
}

// billing orders the actors of rows where the tags were not in billing order.
var billing = map[string][]string{
	"MOV-0679": {"Terence Hill", "Bud Spencer"},
	"MOV-0714": {"Dean Martin", "Jerry Lewis"},
	"MOV-0715": {"Dean Martin", "Jerry Lewis"},
}

// moveActorTags moves the actor tags of every record to its actors column and encodes the
// file again as the data tools do (encoding/csv: minimal quoting, LF). It returns the number
// of records changed.
func moveActorTags(data []byte) ([]byte, int, error) {
	recs, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		return nil, 0, err
	}
	header := recs[0]
	tagsAt, actorsAt, idAt := slices.Index(header, "tags"), slices.Index(header, "actors"), slices.Index(header, "id")
	if tagsAt < 0 || actorsAt < 0 || idAt < 0 {
		return nil, 0, fmt.Errorf("id, tags or actors column missing")
	}
	changed := 0
	for _, rec := range recs[1:] {
		var tags, actors []string
		if rec[actorsAt] != "" {
			actors = strings.Split(rec[actorsAt], "|")
		}
		moved := false
		for _, t := range strings.Split(rec[tagsAt], "|") {
			if name, ok := actorTags[t]; ok {
				moved = true
				if !slices.Contains(actors, name) {
					actors = append(actors, name)
				}
			} else if t != "" {
				tags = append(tags, t)
			}
		}
		if !moved {
			continue
		}
		if order, ok := billing[rec[idAt]]; ok {
			slices.SortStableFunc(actors, func(a, b string) int {
				return rank(order, a) - rank(order, b)
			})
		}
		rec[tagsAt], rec[actorsAt] = strings.Join(tags, "|"), strings.Join(actors, "|")
		changed++
	}
	var out bytes.Buffer
	w := csv.NewWriter(&out)
	if err := w.WriteAll(recs); err != nil {
		return nil, 0, err
	}
	return out.Bytes(), changed, checkMove(data, out.Bytes(), tagsAt, actorsAt)
}

func rank(order []string, name string) int {
	if i := slices.Index(order, name); i >= 0 {
		return i
	}
	return len(order)
}

// checkMove verifies that only tags and actors changed, and that no actor tag is left.
func checkMove(before, after []byte, tagsAt, actorsAt int) error {
	b, _ := csv.NewReader(bytes.NewReader(before)).ReadAll()
	a, err := csv.NewReader(bytes.NewReader(after)).ReadAll()
	if err != nil {
		return err
	}
	if len(a) != len(b) {
		return fmt.Errorf("%d records became %d", len(b), len(a))
	}
	for i := range b {
		for j := range b[i] {
			if j != tagsAt && j != actorsAt && a[i][j] != b[i][j] {
				return fmt.Errorf("record %d, field %d changed", i+1, j+1)
			}
		}
		if i == 0 {
			continue
		}
		for _, t := range strings.Split(a[i][tagsAt], "|") {
			if _, ok := actorTags[t]; ok {
				return fmt.Errorf("record %d still has the tag %s", i+1, t)
			}
		}
	}
	return nil
}
