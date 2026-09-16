package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

// Content failures belong to an utterance/request, never to the connection.
// Only a completed model response is decoded here; transport failures stay errors.
type translationContentError struct{ cause error }

func (e translationContentError) Error() string { return e.cause.Error() }
func decodeModelObject(raw string) M {
	var out M
	if err := json.Unmarshal([]byte(raw), &out); err != nil || out == nil {
		panic(translationContentError{fmt.Errorf("模型返回的内容格式无效；原话已保留。")})
	}
	return out
}

func hasOtherLetters(text string) bool {
	return strings.IndexFunc(text, func(r rune) bool {
		return unicode.IsLetter(r) && !unicode.In(r, unicode.Latin, unicode.Han, unicode.Hiragana, unicode.Katakana)
	}) >= 0
}

// Extract each utterance independently. A long/malformed mixed utterance cannot
// consume the connection retry budget or prevent later utterances from translating.
func planTranslations(segments A) (valid, units A, rejected M) {
	valid, units, rejected = A{}, A{}, M{}
	for _, v := range segments {
		r := obj(v)
		var part A
		if err := attempt(func() { part = translationUnits(A{r}) }); err != nil {
			rejected[str(r["id"])] = err.Error()
		} else {
			valid = append(valid, r)
			units = append(units, part...)
		}
	}
	return
}

// Script alone cannot identify Chinese: 「無料」「大丈夫」must reach the model.
func (l *Live) copyLocalTranscripts(run string) {}

func validExpressionLanguages(x M) bool {
	ja, zh := str(x["japanese"]), str(x["chinese"])
	return strings.TrimSpace(ja) != "" && (kanaRE.MatchString(ja) || hanRE.MatchString(ja) || latinRE.MatchString(ja)) && !hasOtherLetters(ja) && hanRE.MatchString(zh) && !kanaRE.MatchString(zh)
}
