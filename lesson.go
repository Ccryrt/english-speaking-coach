package main

import (
	"encoding/json"
	"path/filepath"
	"time"
)

// ponytail: one micro-lesson per language; add a lesson catalog only after a second lesson is validated.
func lessonState(root string) M {
	return obj(maybeJSON(filepath.Join(root, "Practice", "progress-help.json"), M{"stage": "new", "attempts": A{}}))
}
func lessonStage(state M, day string) string {
	stage := str(state["stage"])
	require(has(stringsA("new", "demonstrate", "supported", "independent", "waiting"), stage), "Invalid lesson state; preserve it and inspect before continuing")
	if stage == "waiting" && str(state["next_review"]) <= day {
		return "retest"
	}
	return stage
}
func lessonView(root, day string) M {
	state := lessonState(root)
	var content M
	must(json.Unmarshal(readAsset("assets/runtime/lesson.json"), &content))
	stage := lessonStage(state, day)
	result := merge(pick(content, "id", "title", "goal", "language", "prompt"), M{
		"stage": stage, "instruction": obj(content["steps"])[stage], "revision": hash(encode(state)),
		"next_review": state["next_review"], "attempt_count": len(arr(state["attempts"])),
	})
	attempts := arr(state["attempts"])
	if len(attempts) > 0 {
		result["previous_context"] = obj(attempts[len(attempts)-1])["context"]
	}
	// Answers and earlier quotes never enter the test page/API response.
	if stage == "demonstrate" || stage == "supported" {
		result["models"] = content["models"]
	}
	if stage == "waiting" {
		attempts := arr(state["attempts"])
		if len(attempts) > 0 {
			result["last_attempt"] = attempts[len(attempts)-1]
		}
	}
	return result
}
func lessonContext(root, day string) M {
	lesson := lessonView(root, day)
	require(lesson["stage"] != "new", "Start guided learning with lesson start first")
	stage := str(lesson["stage"])
	brief := "Guided micro-lesson. Target language: " + str(lesson["language"]) + ". Brief Chinese explanations and help are allowed. One point and one learner turn at a time. Use actual learner facts; never invent their work. " + str(lesson["instruction"])
	if stage == "independent" || stage == "retest" {
		brief += " Do not supply or display an answer, keyword or translation before the attempt. First allow a short unrelated exchange. Ask each of the two goals separately. Help remains available on request; record the actual support for each expression. A newly hidden model is not proof of independent use. For retest use a different task/context from the prior attempt. Never infer pronunciation or unassisted listening from text."
	}
	return M{"lesson": lesson, "voice_brief": brief, "next_action": "Open the guided page in the right panel. Follow lesson.stage and references/guided-learning.md; wait for the learner after each short turn. Do not begin a free scene or auto-advance stages."}
}
func lessonCommand(root, day, action string, args M) M {
	if action == "show" {
		return lessonView(root, day)
	}
	require(has(stringsA("start", "next", "finish"), action), "Unknown lesson action")
	withLock(filepath.Join(root, ".write.lock"), func() {
		state := lessonState(root)
		stage := lessonStage(state, day)
		if action == "start" && stage != "new" {
			return
		} // Resume without erasing learning evidence.
		require(str(args["expected-revision"]) == hash(encode(state)), "Lesson changed; read lesson show and retry with its expected revision")
		switch action {
		case "start":
			state["stage"] = "demonstrate"
		case "next":
			require(stage == "demonstrate" || stage == "supported", "Only demonstrated/supported practice can advance; independent attempts need evidence")
			state["stage"] = "supported"
			if stage == "supported" {
				state["stage"] = "independent"
			}
		case "finish":
			require(stage == "independent" || stage == "retest", "Finish requires an independent attempt or a due retest")
			require(str(args["input"]) != "", "Finish needs actual learner evidence")
			evidence := obj(readJSON(str(args["input"])))
			require(exactKeys(evidence, "context", "source", "responses"), "Evidence needs context, source and responses")
			shortText(evidence["context"], "Different task/context", 500)
			source := obj(evidence["source"])
			require(exactKeys(source, "kind", "reference"), "Source needs kind and reference")
			require(has(stringsA("text", "voice_transcript"), source["kind"]), "Use text or voice_transcript evidence; no inferred audio score")
			shortText(source["reference"], "Source task/turn reference", 500)
			responses := arr(evidence["responses"])
			require(len(responses) == 2, "Observe both progress and help separately")
			independent := true
			seen := M{}
			for _, v := range responses {
				response := obj(v)
				require(exactKeys(response, "id", "quote", "support", "success"), "Response needs id, quote, support, success")
				id := str(response["id"])
				require((id == "progress" || id == "help") && seen[id] == nil, "Each goal needs exactly one response")
				seen[id] = true
				shortText(response["quote"], "Actual learner quote (or observed no response)", 1000)
				require(has(stringsA("none", "keyword", "model"), response["support"]), "Unknown support level")
				_, valid := response["success"].(bool)
				require(valid, "Success must be a boolean")
				independent = independent && truth(response["success"]) && response["support"] == "none"
			}
			attempts := arr(state["attempts"])
			if len(attempts) > 0 {
				last := obj(attempts[len(attempts)-1])
				require(day > str(last["date"]), "Retest must happen on a later date")
				require(str(evidence["context"]) != str(last["context"]), "Retest needs a different task/context")
			}
			attempt := merge(evidence, M{"date": day, "stage": stage, "independent": independent})
			state["attempts"] = append(attempts, attempt)
			interval := 1
			if stage == "retest" && independent {
				interval = 7
			}
			date, err := time.Parse("2006-01-02", day)
			must(err)
			state["stage"], state["next_review"] = "waiting", date.AddDate(0, 0, interval).Format("2006-01-02")
		}
		state["updated"] = now()
		writeJSON(filepath.Join(root, "Practice", "progress-help.json"), state)
	})
	return lessonView(root, day)
}
