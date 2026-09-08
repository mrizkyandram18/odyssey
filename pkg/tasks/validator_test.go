package tasks

import (
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
