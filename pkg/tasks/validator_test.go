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
