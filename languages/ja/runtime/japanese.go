package main

import "strings"

// Kana is optional support, never evidence of the learner's pronunciation.
func validKanaReading(text string) bool {
	return strings.TrimSpace(text) != "" && len([]rune(text)) <= 4000 && kanaRE.MatchString(text) && !hanRE.MatchString(text) && !hasOtherLetters(text)
}

func validateJapaneseNotes(expression M) {
	if reading := str(expression["reading"]); reading != "" {
		require(validKanaReading(reading), "读音需使用假名，可保留数字和英文缩写；不使用罗马音代替假名")
	}
	if value := expression["register"]; value != nil {
		shortText(value, "适用对象与礼貌", 400)
	}
	if value := expression["correction_kind"]; value != nil {
		require(has(stringsA("error", "naturalness", "register", "already_correct"), value), "无效的表达建议类型")
	}
}
