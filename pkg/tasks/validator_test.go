package tasks

import (
	"encoding/json"
	"math"
	"testing"
)

func TestEngine_ValidateTaskInput(t *testing.T) {
	engine := NewEngine()

	// 1. Valid Canonical VIDEO
	t.Run("Valid VIDEO task", func(t *testing.T) {
		task := &TaskInput{
			Title:       "Watch Video",
			TaskType:    "VIDEO",
			RewardCoins: 50,
			RewardXP:    100,
			StepOrder:   1,
			Config: map[string]any{
				"youtube_url": "https://www.youtube.com/watch?v=123",
			},
		}
		if err := engine.ValidateTaskInput(task); err != nil {
			t.Fatalf("expected valid video task, got: %v", err)
		}
	})

	// 2. Invalid Video URL (non-http/https)
	t.Run("Invalid Video URL rejected", func(t *testing.T) {
		task := &TaskInput{
			Title:       "Watch Video",
			TaskType:    "VIDEO",
			RewardCoins: 50,
			RewardXP:    100,
			Config: map[string]any{
				"youtube_url": "javascript:alert(1)",
			},
		}
		if err := engine.ValidateTaskInput(task); err == nil {
			t.Fatalf("expected error on malicious video URL, got nil")
		}
	})

	// 3. Valid Composite VIDEO + QUIZ
	t.Run("Valid Composite VIDEO + QUIZ task", func(t *testing.T) {
		task := &TaskInput{
			Title:       "Video Lesson with Quiz",
			TaskType:    "VIDEO",
			RewardCoins: 75,
			RewardXP:    150,
			Config: map[string]any{
				"youtube_url": "https://www.youtube.com/watch?v=abc",
				"questions": []any{
					map[string]any{
						"id":             "1",
						"question":       "What was the main topic?",
						"options":        []any{"Topic A", "Topic B"},
						"correct_answer": "Topic A",
					},
				},
			},
		}
		if err := engine.ValidateTaskInput(task); err != nil {
			t.Fatalf("expected valid composite task, got: %v", err)
		}
	})

	// 4. Composite with Duplicate Question IDs Rejected
	t.Run("Composite with duplicate question IDs rejected", func(t *testing.T) {
		task := &TaskInput{
			Title:       "Video Lesson with Duplicate Questions",
			TaskType:    "VIDEO",
			RewardCoins: 75,
			RewardXP:    150,
			Config: map[string]any{
				"youtube_url": "https://www.youtube.com/watch?v=abc",
				"questions": []any{
					map[string]any{"id": "q1", "question": "Q1", "options": []any{"A", "B"}, "correct_answer": "A"},
					map[string]any{"id": "q1", "question": "Q2", "options": []any{"A", "B"}, "correct_answer": "B"},
				},
			},
		}
		if err := engine.ValidateTaskInput(task); err == nil {
			t.Fatalf("expected error on duplicate question ID, got nil")
		}
	})

	// 5. Valid Composite DOCUMENT + TEXT
	t.Run("Valid Composite DOCUMENT + TEXT task", func(t *testing.T) {
		task := &TaskInput{
			Title:       "Homework with Reflection",
			TaskType:    "DOCUMENT_UPLOAD",
			RewardCoins: 100,
			Config: map[string]any{
				"max_file_size_mb":   float64(10),
				"minimum_characters": float64(20),
				"maximum_characters": float64(500),
			},
		}
		if err := engine.ValidateTaskInput(task); err != nil {
			t.Fatalf("expected valid composite document+text task, got: %v", err)
		}
	})

	// 6. Unknown Task Type Rejected
	t.Run("Unknown task type rejected", func(t *testing.T) {
		task := &TaskInput{
			Title:    "Mystery Task",
			TaskType: "CRYPTO_MINING",
		}
		if err := engine.ValidateTaskInput(task); err == nil {
			t.Fatalf("expected error on unknown task type, got nil")
		}
	})

	// 7. Negative Reward Rejected
	t.Run("Negative reward coins rejected", func(t *testing.T) {
		task := &TaskInput{
			Title:       "Bad Reward Task",
			TaskType:    "TEXT_RESPONSE",
			RewardCoins: -10,
		}
		if err := engine.ValidateTaskInput(task); err == nil {
			t.Fatalf("expected error on negative reward coins, got nil")
		}
	})
}

func TestResolveEvaluationTypeForConfig(t *testing.T) {
	cases := []struct {
		name     string
		taskType string
		explicit string
		config   map[string]any
		want     string
	}{
		// AC2: VIDEO + essay prompt forces ADMIN_REVIEW
		{"VIDEO essay defaults to ADMIN_REVIEW", "VIDEO", "", map[string]any{"video_url": "https://www.youtube.com/watch?v=x", "prompt": "Jelaskan..."}, "ADMIN_REVIEW"},
		// Backward compat: legacy stored AUTO is overridden by essay prompt
		{"VIDEO essay overrides explicit AUTO", "VIDEO", "AUTO", map[string]any{"prompt": "Jelaskan..."}, "ADMIN_REVIEW"},
		{"VIDEO essay lowercase type", "video", "", map[string]any{"prompt": "Jelaskan..."}, "ADMIN_REVIEW"},
		// AC1: VIDEO without prompt keeps existing behavior
		{"VIDEO watch-only defaults to AUTO", "VIDEO", "", map[string]any{"video_url": "https://www.youtube.com/watch?v=x"}, "AUTO"},
		{"VIDEO watch-only explicit AUTO stays AUTO", "VIDEO", "AUTO", nil, "AUTO"},
		{"VIDEO blank prompt stays AUTO", "VIDEO", "", map[string]any{"prompt": "   "}, "AUTO"},
		{"VIDEO quiz stays AUTO", "VIDEO", "", map[string]any{
			"youtube_url": "https://www.youtube.com/watch?v=x",
			"questions": []any{
				map[string]any{"id": "1", "question": "Q?", "options": []any{"A", "B"}, "correct_answer": "A"},
			},
		}, "AUTO"},
		// Existing types unchanged
		{"TEXT_RESPONSE stays ADMIN_REVIEW", "TEXT_RESPONSE", "", map[string]any{"prompt": "Ceritakan..."}, "ADMIN_REVIEW"},
		{"MINI_GAME stays AUTO", "MINI_GAME", "", nil, "AUTO"},
		{"explicit ADMIN_REVIEW preserved", "QUIZ", "ADMIN_REVIEW", nil, "ADMIN_REVIEW"},
		// VIDEO + recording forces ADMIN_REVIEW (even legacy stored AUTO)
		{"VIDEO recording defaults to ADMIN_REVIEW", "VIDEO", "", map[string]any{"recording": map[string]any{"enabled": true, "max_duration_seconds": float64(60)}}, "ADMIN_REVIEW"},
		{"VIDEO recording overrides explicit AUTO", "VIDEO", "AUTO", map[string]any{"recording": map[string]any{"enabled": true}}, "ADMIN_REVIEW"},
		{"VIDEO recording disabled stays AUTO", "VIDEO", "", map[string]any{"recording": map[string]any{"enabled": false}}, "AUTO"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveEvaluationTypeForConfig(tc.taskType, tc.explicit, tc.config); got != tc.want {
				t.Errorf("ResolveEvaluationTypeForConfig(%q, %q, %v) = %q, want %q",
					tc.taskType, tc.explicit, tc.config, got, tc.want)
			}
		})
	}
}

func decisionTestConfig() map[string]any {
	return map[string]any{
		"game":         "DECISION_FINANCE",
		"target_score": float64(100),
		"scenario": map[string]any{
			"initial_balance": float64(500000),
			"events": []any{
				map[string]any{
					"id":    "day_1",
					"title": "Nongkrong",
					"options": []any{
						map[string]any{"id": "a", "label": "Ikut", "delta": float64(-75000)},
						map[string]any{"id": "b", "label": "Tolak", "delta": float64(0)},
					},
				},
				map[string]any{
					"id":    "day_2",
					"title": "Kuota",
					"options": []any{
						map[string]any{"id": "a", "label": "Paket 50rb", "delta": float64(-50000)},
						map[string]any{"id": "b", "label": "WiFi gratis", "delta": float64(0)},
					},
				},
			},
		},
	}
}

func TestDecisionScenarioValidation(t *testing.T) {
	engine := NewEngine()

	t.Run("Valid MINI_GAME decision scenario accepted", func(t *testing.T) {
		task := &TaskInput{Title: "Bos Keuanganmu", TaskType: "MINI_GAME", RewardCoins: 50, RewardXP: 100, Config: decisionTestConfig()}
		if err := engine.ValidateTaskInput(task); err != nil {
			t.Fatalf("expected valid scenario, got: %v", err)
		}
	})

	t.Run("Legacy MINI_GAME without scenario still accepted", func(t *testing.T) {
		task := &TaskInput{Title: "Memory", TaskType: "MINI_GAME", Config: map[string]any{"game": "MEMORY_MATCH", "target_score": float64(80)}}
		if err := engine.ValidateTaskInput(task); err != nil {
			t.Fatalf("expected legacy game accepted, got: %v", err)
		}
	})

	t.Run("Scenario without events rejected", func(t *testing.T) {
		cfg := decisionTestConfig()
		cfg["scenario"] = map[string]any{"initial_balance": float64(500000)}
		task := &TaskInput{Title: "Bad", TaskType: "MINI_GAME", Config: cfg}
		if err := engine.ValidateTaskInput(task); err == nil {
			t.Fatalf("expected error on missing events, got nil")
		}
	})

	t.Run("Scenario with single-option event rejected", func(t *testing.T) {
		cfg := decisionTestConfig()
		ev := cfg["scenario"].(map[string]any)["events"].([]any)[0].(map[string]any)
		ev["options"] = []any{map[string]any{"id": "a", "label": "Only", "delta": float64(1)}}
		task := &TaskInput{Title: "Bad", TaskType: "MINI_GAME", Config: cfg}
		if err := engine.ValidateTaskInput(task); err == nil {
			t.Fatalf("expected error on single option, got nil")
		}
	})

	t.Run("VIDEO recording config accepted without youtube_url", func(t *testing.T) {
		task := &TaskInput{Title: "Self intro", TaskType: "VIDEO", Config: map[string]any{
			"recording": map[string]any{"enabled": true, "max_duration_seconds": float64(60), "camera_facing": "user"},
		}}
		if err := engine.ValidateTaskInput(task); err != nil {
			t.Fatalf("expected recording config accepted, got: %v", err)
		}
	})

	t.Run("VIDEO recording invalid duration rejected", func(t *testing.T) {
		task := &TaskInput{Title: "Self intro", TaskType: "VIDEO", Config: map[string]any{
			"recording": map[string]any{"enabled": true, "max_duration_seconds": float64(9999)},
		}}
		if err := engine.ValidateTaskInput(task); err == nil {
			t.Fatalf("expected error on invalid duration, got nil")
		}
	})
}

func TestValidateDecisionChoices(t *testing.T) {
	cfg := decisionTestConfig()

	t.Run("Valid complete path computes balance", func(t *testing.T) {
		bal, err := ValidateDecisionChoices(cfg, map[string]any{"day_1": "a", "day_2": "b"})
		if err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
		if bal != 425000 {
			t.Fatalf("expected 425000, got %d", bal)
		}
	})

	t.Run("All-saving path keeps initial balance", func(t *testing.T) {
		bal, err := ValidateDecisionChoices(cfg, map[string]any{"day_1": "b", "day_2": "b"})
		if err != nil || bal != 500000 {
			t.Fatalf("expected 500000 nil error, got %d %v", bal, err)
		}
	})

	t.Run("Missing event rejected", func(t *testing.T) {
		if _, err := ValidateDecisionChoices(cfg, map[string]any{"day_1": "a"}); err == nil {
			t.Fatalf("expected error on missing event, got nil")
		}
	})

	t.Run("Invalid option rejected (score spoof attempt)", func(t *testing.T) {
		if _, err := ValidateDecisionChoices(cfg, map[string]any{"day_1": "zzz", "day_2": "b"}); err == nil {
			t.Fatalf("expected error on invalid option, got nil")
		}
	})

	t.Run("Extra unknown event rejected", func(t *testing.T) {
		if _, err := ValidateDecisionChoices(cfg, map[string]any{"day_1": "a", "day_2": "b", "day_99": "a"}); err == nil {
			t.Fatalf("expected error on extra event, got nil")
		}
	})

	t.Run("Nil choices rejected", func(t *testing.T) {
		if _, err := ValidateDecisionChoices(cfg, nil); err == nil {
			t.Fatalf("expected error on nil choices, got nil")
		}
	})

	t.Run("Non-scenario config rejected", func(t *testing.T) {
		if _, err := ValidateDecisionChoices(map[string]any{"game": "MEMORY"}, map[string]any{}); err == nil {
			t.Fatalf("expected error on non-scenario config, got nil")
		}
	})
}

func decisionPriorityTestConfig() map[string]any {
	return map[string]any{
		"game":         "DECISION_PRIORITY",
		"target_score": float64(100),
		"scenario": map[string]any{
			"initial_balance": float64(0),
			"currency":        "POINTS",
			"events": []any{
				map[string]any{
					"id":    "ev_1",
					"title": "Kebutuhan Pokok - Beras",
					"options": []any{
						map[string]any{"id": "a", "label": "Beras medium", "delta": float64(20)},
						map[string]any{"id": "b", "label": "Beras premium", "delta": float64(5)},
						map[string]any{"id": "c", "label": "Mie instan", "delta": float64(0)},
					},
				},
				map[string]any{
					"id":    "ev_2",
					"title": "Lauk Pauk - Protein",
					"options": []any{
						map[string]any{"id": "a", "label": "Telur dan tempe", "delta": float64(20)},
						map[string]any{"id": "b", "label": "Daging ayam", "delta": float64(5)},
						map[string]any{"id": "c", "label": "Sosis", "delta": float64(0)},
					},
				},
				map[string]any{
					"id":    "ev_3",
					"title": "Sayuran Segar",
					"options": []any{
						map[string]any{"id": "a", "label": "Bayam dan wortel", "delta": float64(20)},
						map[string]any{"id": "b", "label": "Sayur kalengan", "delta": float64(5)},
						map[string]any{"id": "c", "label": "Lewatkan sayur", "delta": float64(0)},
					},
				},
				map[string]any{
					"id":    "ev_4",
					"title": "Bumbu Dapur",
					"options": []any{
						map[string]any{"id": "a", "label": "Paket bumbu", "delta": float64(20)},
						map[string]any{"id": "b", "label": "Bumbu botol", "delta": float64(5)},
						map[string]any{"id": "c", "label": "Pakai garam", "delta": float64(0)},
					},
				},
				map[string]any{
					"id":    "ev_5",
					"title": "Minyak Goreng",
					"options": []any{
						map[string]any{"id": "a", "label": "Minyak pouch", "delta": float64(20)},
						map[string]any{"id": "b", "label": "Minyak jagung", "delta": float64(0)},
						map[string]any{"id": "c", "label": "Minta tetangga", "delta": float64(5)},
					},
				},
				map[string]any{
					"id":    "ev_6",
					"title": "Sisa Anggaran",
					"options": []any{
						map[string]any{"id": "a", "label": "Simpan tabungan", "delta": float64(20)},
						map[string]any{"id": "b", "label": "Permen", "delta": float64(5)},
						map[string]any{"id": "c", "label": "Hutang", "delta": float64(0)},
					},
				},
			},
		},
	}
}

func TestParseStrictPositiveInt32(t *testing.T) {
	t.Run("Valid int", func(t *testing.T) {
		val, err := ParseStrictPositiveInt32(100)
		if err != nil || val != 100 {
			t.Fatalf("expected 100, got %d %v", val, err)
		}
	})

	t.Run("Valid int32", func(t *testing.T) {
		val, err := ParseStrictPositiveInt32(int32(100))
		if err != nil || val != 100 {
			t.Fatalf("expected 100, got %d %v", val, err)
		}
	})

	t.Run("Valid int64", func(t *testing.T) {
		val, err := ParseStrictPositiveInt32(int64(100))
		if err != nil || val != 100 {
			t.Fatalf("expected 100, got %d %v", val, err)
		}
	})

	t.Run("Valid float64 without fraction", func(t *testing.T) {
		val, err := ParseStrictPositiveInt32(float64(100))
		if err != nil || val != 100 {
			t.Fatalf("expected 100, got %d %v", val, err)
		}
	})

	t.Run("Valid json.Number", func(t *testing.T) {
		val, err := ParseStrictPositiveInt32(json.Number("100"))
		if err != nil || val != 100 {
			t.Fatalf("expected 100, got %d %v", val, err)
		}
	})

	t.Run("Boundary max int32", func(t *testing.T) {
		val, err := ParseStrictPositiveInt32(math.MaxInt32)
		if err != nil || val != math.MaxInt32 {
			t.Fatalf("expected MaxInt32, got %d %v", val, err)
		}
	})

	t.Run("Nil rejected", func(t *testing.T) {
		if _, err := ParseStrictPositiveInt32(nil); err == nil {
			t.Fatalf("expected error on nil")
		}
	})

	t.Run("String rejected", func(t *testing.T) {
		if _, err := ParseStrictPositiveInt32("100"); err == nil {
			t.Fatalf("expected error on string")
		}
	})

	t.Run("Boolean rejected", func(t *testing.T) {
		if _, err := ParseStrictPositiveInt32(true); err == nil {
			t.Fatalf("expected error on bool")
		}
	})

	t.Run("Float with decimal rejected", func(t *testing.T) {
		if _, err := ParseStrictPositiveInt32(100.5); err == nil {
			t.Fatalf("expected error on float with fraction")
		}
	})

	t.Run("Float NaN rejected", func(t *testing.T) {
		if _, err := ParseStrictPositiveInt32(math.NaN()); err == nil {
			t.Fatalf("expected error on NaN")
		}
	})

	t.Run("Float Inf rejected", func(t *testing.T) {
		if _, err := ParseStrictPositiveInt32(math.Inf(1)); err == nil {
			t.Fatalf("expected error on +Inf")
		}
	})

	t.Run("Zero rejected", func(t *testing.T) {
		if _, err := ParseStrictPositiveInt32(0); err == nil {
			t.Fatalf("expected error on 0")
		}
	})

	t.Run("Negative int rejected", func(t *testing.T) {
		if _, err := ParseStrictPositiveInt32(-10); err == nil {
			t.Fatalf("expected error on negative int")
		}
	})

	t.Run("Negative float rejected", func(t *testing.T) {
		if _, err := ParseStrictPositiveInt32(-100.0); err == nil {
			t.Fatalf("expected error on negative float")
		}
	})

	t.Run("Int64 overflow rejected", func(t *testing.T) {
		if _, err := ParseStrictPositiveInt32(int64(math.MaxInt32) + 1); err == nil {
			t.Fatalf("expected error on int64 overflow")
		}
	})

	t.Run("Float64 overflow rejected", func(t *testing.T) {
		if _, err := ParseStrictPositiveInt32(float64(math.MaxInt32) + 1000); err == nil {
			t.Fatalf("expected error on float64 overflow")
		}
	})
}

func TestSafeAddInt32(t *testing.T) {
	t.Run("Normal addition", func(t *testing.T) {
		sum, err := SafeAddInt32(10, 20)
		if err != nil || sum != 30 {
			t.Fatalf("expected 30, got %d %v", sum, err)
		}
	})

	t.Run("Negative delta", func(t *testing.T) {
		sum, err := SafeAddInt32(100, -20)
		if err != nil || sum != 80 {
			t.Fatalf("expected 80, got %d %v", sum, err)
		}
	})

	t.Run("Upper boundary safe", func(t *testing.T) {
		sum, err := SafeAddInt32(math.MaxInt32-10, 10)
		if err != nil || sum != math.MaxInt32 {
			t.Fatalf("expected MaxInt32, got %d %v", sum, err)
		}
	})

	t.Run("Lower boundary safe", func(t *testing.T) {
		sum, err := SafeAddInt32(math.MinInt32+10, -10)
		if err != nil || sum != math.MinInt32 {
			t.Fatalf("expected MinInt32, got %d %v", sum, err)
		}
	})

	t.Run("Overflow detected", func(t *testing.T) {
		if _, err := SafeAddInt32(math.MaxInt32, 1); err == nil {
			t.Fatalf("expected overflow error")
		}
	})

	t.Run("Underflow detected", func(t *testing.T) {
		if _, err := SafeAddInt32(math.MinInt32, -1); err == nil {
			t.Fatalf("expected underflow error")
		}
	})

	t.Run("Operand a out of range", func(t *testing.T) {
		if _, err := SafeAddInt32(math.MaxInt32+1, 0); err == nil {
			t.Fatalf("expected error on operand a")
		}
	})
}

func TestEvaluateDecisionPriorityScore(t *testing.T) {
	cfg := decisionPriorityTestConfig()

	t.Run("Optimal choices achieve 120 points (passed)", func(t *testing.T) {
		choices := map[string]any{
			"ev_1": "a",
			"ev_2": "a",
			"ev_3": "a",
			"ev_4": "a",
			"ev_5": "a",
			"ev_6": "a",
		}
		score, bal, err := EvaluateDecisionPriorityScore(cfg, choices)
		if err != nil {
			t.Fatalf("expected passing score, got: %v", err)
		}
		if score != 120 || bal != 120 {
			t.Fatalf("expected score 120 and balance 120, got score %d bal %d", score, bal)
		}
	})

	t.Run("Exact passing choices achieve 100 points (passed)", func(t *testing.T) {
		choices := map[string]any{
			"ev_1": "a", // 20
			"ev_2": "a", // 20
			"ev_3": "a", // 20
			"ev_4": "a", // 20
			"ev_5": "a", // 20
			"ev_6": "c", // 0
		}
		score, bal, err := EvaluateDecisionPriorityScore(cfg, choices)
		if err != nil {
			t.Fatalf("expected passing score, got: %v", err)
		}
		if score != 100 || bal != 100 {
			t.Fatalf("expected score 100 and balance 100, got score %d bal %d", score, bal)
		}
	})

	t.Run("Failing choices achieve 90 points (rejected)", func(t *testing.T) {
		choices := map[string]any{
			"ev_1": "a", // 20
			"ev_2": "a", // 20
			"ev_3": "a", // 20
			"ev_4": "a", // 20
			"ev_5": "c", // 5
			"ev_6": "b", // 5
		}
		_, _, err := EvaluateDecisionPriorityScore(cfg, choices)
		if err == nil {
			t.Fatalf("expected error for score 90 < target 100, got nil")
		}
	})

	t.Run("Worst choices achieve 0 points (rejected)", func(t *testing.T) {
		choices := map[string]any{
			"ev_1": "c", // 0
			"ev_2": "c", // 0
			"ev_3": "c", // 0
			"ev_4": "c", // 0
			"ev_5": "b", // 0
			"ev_6": "c", // 0
		}
		_, _, err := EvaluateDecisionPriorityScore(cfg, choices)
		if err == nil {
			t.Fatalf("expected error for score 0 < target 100, got nil")
		}
	})

	t.Run("Missing scenario currency rejected", func(t *testing.T) {
		badCfg := decisionPriorityTestConfig()
		delete(badCfg["scenario"].(map[string]any), "currency")
		choices := map[string]any{"ev_1": "a", "ev_2": "a", "ev_3": "a", "ev_4": "a", "ev_5": "a", "ev_6": "a"}
		if _, _, err := EvaluateDecisionPriorityScore(badCfg, choices); err == nil {
			t.Fatalf("expected error on missing currency")
		}
	})

	t.Run("Wrong currency rejected", func(t *testing.T) {
		badCfg := decisionPriorityTestConfig()
		badCfg["scenario"].(map[string]any)["currency"] = "IDR"
		choices := map[string]any{"ev_1": "a", "ev_2": "a", "ev_3": "a", "ev_4": "a", "ev_5": "a", "ev_6": "a"}
		if _, _, err := EvaluateDecisionPriorityScore(badCfg, choices); err == nil {
			t.Fatalf("expected error on wrong currency")
		}
	})

	t.Run("Wrong game name rejected", func(t *testing.T) {
		badCfg := decisionPriorityTestConfig()
		badCfg["game"] = "DECISION_FINANCE"
		choices := map[string]any{"ev_1": "a", "ev_2": "a", "ev_3": "a", "ev_4": "a", "ev_5": "a", "ev_6": "a"}
		if _, _, err := EvaluateDecisionPriorityScore(badCfg, choices); err == nil {
			t.Fatalf("expected error on wrong game name")
		}
	})

	t.Run("Invalid target_score rejected", func(t *testing.T) {
		badCfg := decisionPriorityTestConfig()
		badCfg["target_score"] = 100.5
		choices := map[string]any{"ev_1": "a", "ev_2": "a", "ev_3": "a", "ev_4": "a", "ev_5": "a", "ev_6": "a"}
		if _, _, err := EvaluateDecisionPriorityScore(badCfg, choices); err == nil {
			t.Fatalf("expected error on invalid target_score")
		}
	})

	t.Run("Delta overflow attempt in event options rejected", func(t *testing.T) {
		badCfg := decisionPriorityTestConfig()
		badCfg["scenario"].(map[string]any)["events"].([]any)[0].(map[string]any)["options"].([]any)[0].(map[string]any)["delta"] = float64(math.MaxInt32)
		badCfg["scenario"].(map[string]any)["events"].([]any)[1].(map[string]any)["options"].([]any)[0].(map[string]any)["delta"] = float64(10)
		choices := map[string]any{"ev_1": "a", "ev_2": "a", "ev_3": "a", "ev_4": "a", "ev_5": "a", "ev_6": "a"}
		if _, _, err := EvaluateDecisionPriorityScore(badCfg, choices); err == nil {
			t.Fatalf("expected overflow error on delta addition")
		}
	})

	t.Run("Incomplete choices rejected", func(t *testing.T) {
		choices := map[string]any{"ev_1": "a", "ev_2": "a"}
		if _, _, err := EvaluateDecisionPriorityScore(cfg, choices); err == nil {
			t.Fatalf("expected error on incomplete choices")
		}
	})

	t.Run("Invalid option ID rejected", func(t *testing.T) {
		choices := map[string]any{"ev_1": "xyz", "ev_2": "a", "ev_3": "a", "ev_4": "a", "ev_5": "a", "ev_6": "a"}
		if _, _, err := EvaluateDecisionPriorityScore(cfg, choices); err == nil {
			t.Fatalf("expected error on invalid option ID")
		}
	})

	t.Run("Nil choices rejected", func(t *testing.T) {
		if _, _, err := EvaluateDecisionPriorityScore(cfg, nil); err == nil {
			t.Fatalf("expected error on nil choices")
		}
	})

	t.Run("Nil config rejected", func(t *testing.T) {
		if _, _, err := EvaluateDecisionPriorityScore(nil, map[string]any{}); err == nil {
			t.Fatalf("expected error on nil config")
		}
	})
}
