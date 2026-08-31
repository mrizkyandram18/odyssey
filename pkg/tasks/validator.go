package tasks

import (
	"fmt"
	"strings"
)

// TaskInput represents the generic task creation/update payload.
type TaskInput struct {
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	TaskType       string         `json:"task_type"`
	EvaluationType string         `json:"evaluation_type,omitempty"`
	StepOrder      int            `json:"step_order"`
	ActiveDate     string         `json:"active_date,omitempty"`
	RewardCoins    int            `json:"reward_coins"`
	RewardXP       int            `json:"reward_xp"`
	Config         map[string]any `json:"config"`
	IsActive       bool           `json:"is_active"`
}

// CapabilityValidator validates a single task capability configuration.
type CapabilityValidator func(config map[string]any) error

// Engine manages capability-based task validation and execution metadata.
type Engine struct {
	validators map[string]CapabilityValidator
}

// NewEngine creates a Task Engine initialized with default capability validators.
func NewEngine() *Engine {
	e := &Engine{
		validators: make(map[string]CapabilityValidator),
	}
	e.RegisterDefaultValidators()
	return e
}

// RegisterCapability registers a capability validator.
func (e *Engine) RegisterCapability(capability string, validator CapabilityValidator) {
	e.validators[strings.ToLower(capability)] = validator
}

// RegisterDefaultValidators registers the canonical capability validators.
func (e *Engine) RegisterDefaultValidators() {
	// 1. Video Capability
	e.RegisterCapability("video", func(cfg map[string]any) error {
		urlVal, _ := cfg["video_url"].(string)
		if urlVal == "" {
			urlVal, _ = cfg["youtube_url"].(string)
		}
		if urlVal != "" && !strings.HasPrefix(urlVal, "http://") && !strings.HasPrefix(urlVal, "https://") {
			return fmt.Errorf("URL video harus diawali http:// atau https://")
		}
		if dur, ok := cfg["minimum_duration_seconds"].(float64); ok && dur < 0 {
			return fmt.Errorf("minimum_duration_seconds tidak boleh negatif")
		}
		return nil
	})

	// 2. Quiz Capability
	e.RegisterCapability("quiz", func(cfg map[string]any) error {
		questionsRaw, ok := cfg["questions"]
		if !ok || questionsRaw == nil {
			return fmt.Errorf("tugas kuis wajib memiliki minimal 1 pertanyaan di config.questions")
		}
		questions, ok := questionsRaw.([]any)
		if !ok || len(questions) == 0 {
			return fmt.Errorf("tugas kuis wajib memiliki minimal 1 pertanyaan di config.questions")
		}
		if len(questions) > 50 {
			return fmt.Errorf("jumlah soal kuis maksimal 50 pertanyaan")
		}

		seenIDs := make(map[string]bool)
		for i, qRaw := range questions {
			qMap, ok := qRaw.(map[string]any)
			if !ok {
				return fmt.Errorf("soal kuis #%d tidak valid", i+1)
			}
			qID := fmt.Sprintf("%v", qMap["id"])
			if strings.TrimSpace(qID) == "" || qID == "<nil>" {
				return fmt.Errorf("soal kuis #%d wajib memiliki id", i+1)
			}
			if seenIDs[qID] {
				return fmt.Errorf("soal kuis #%d memiliki id duplikat: %s", i+1, qID)
			}
			seenIDs[qID] = true

			qText, _ := qMap["question"].(string)
			if strings.TrimSpace(qText) == "" {
				return fmt.Errorf("soal kuis #%d wajib memiliki teks pertanyaan", i+1)
			}
			correct := fmt.Sprintf("%v", qMap["correct_answer"])
			if strings.TrimSpace(correct) == "" || correct == "<nil>" {
				return fmt.Errorf("soal kuis #%d wajib memiliki kunci jawaban (correct_answer)", i+1)
			}
			opts, ok := qMap["options"].([]any)
			if !ok || len(opts) < 2 {
				return fmt.Errorf("soal kuis #%d wajib memiliki minimal 2 pilihan jawaban (options)", i+1)
			}
			if len(opts) > 10 {
				return fmt.Errorf("soal kuis #%d memiliki terlalu banyak pilihan jawaban (maksimal 10)", i+1)
			}
		}
		return nil
	})

	// 3. Photo Capability
	e.RegisterCapability("photo", func(cfg map[string]any) error {
		if mf, ok := cfg["max_files"].(float64); ok && (mf <= 0 || mf > 10) {
			return fmt.Errorf("max_files harus antara 1 dan 10")
		}
		return nil
	})

	// 4. Document Capability
	e.RegisterCapability("document", func(cfg map[string]any) error {
		if ms, ok := cfg["max_file_size_mb"].(float64); ok && (ms <= 0 || ms > 25) {
			return fmt.Errorf("max_file_size_mb harus antara 1 dan 25 MB")
		}
		if attachURL, ok := cfg["attachment_url"].(string); ok && attachURL != "" {
			if !strings.HasPrefix(attachURL, "http://") && !strings.HasPrefix(attachURL, "https://") {
				return fmt.Errorf("attachment_url harus berupa URL valid (http/https)")
			}
		}
		return nil
	})

	// 5. Text Response Capability
	e.RegisterCapability("text", func(cfg map[string]any) error {
		minChars := 0
		maxChars := 5000
		if mc, ok := cfg["minimum_characters"].(float64); ok {
			minChars = int(mc)
		}
		if mc, ok := cfg["maximum_characters"].(float64); ok {
			maxChars = int(mc)
		}
		if minChars < 0 || maxChars <= 0 || minChars > maxChars || maxChars > 10000 {
			return fmt.Errorf("batasan karakter tidak valid (min: %d, max: %d)", minChars, maxChars)
		}
		return nil
	})

	// 6. Mini Game Capability
	e.RegisterCapability("game", func(cfg map[string]any) error {
		if ts, ok := cfg["target_score"].(float64); ok && (ts < 0 || ts > 1000000) {
			return fmt.Errorf("target_score harus antara 0 dan 1,000,000")
		}
		return nil
	})

	// 7. Checklist Capability
	e.RegisterCapability("checklist", func(cfg map[string]any) error {
		itemsRaw, ok := cfg["items"]
		if ok && itemsRaw != nil {
			items, ok := itemsRaw.([]any)
			if !ok || len(items) == 0 {
				return fmt.Errorf("checklist harus memiliki minimal 1 item")
			}
			if len(items) > 50 {
				return fmt.Errorf("checklist memiliki terlalu banyak item (maksimal 50)")
			}
		}
		return nil
	})
}

// Canonical task types mapping to underlying capabilities
var taskTypeCapabilities = map[string][]string{
	"VIDEO":           {"video"},
	"QUIZ":            {"quiz"},
	"PHOTO_UPLOAD":    {"photo"},
	"DOCUMENT_UPLOAD": {"document"},
	"TEXT_RESPONSE":   {"text"},
	"MINI_GAME":       {"game"},
	"CHECKLIST":       {"checklist"},
	// Compatibility aliases mapping directly to canonical capabilities
	"VIDEO_QUIZ":    {"video", "quiz"},
	"PHOTO_PROOF":   {"photo"},
	"YOUTUBE_VIDEO": {"video"},
	"GENERAL":       {},
}

// ValidateTaskInput performs generic and capability-specific validation.
func (e *Engine) ValidateTaskInput(req *TaskInput) error {
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("judul tugas tidak boleh kosong")
	}
	if len(req.Title) > 255 {
		return fmt.Errorf("judul tugas terlalu panjang (maksimal 255 karakter)")
	}

	normalizedType := strings.ToUpper(strings.TrimSpace(req.TaskType))
	caps, isKnown := taskTypeCapabilities[normalizedType]
	if !isKnown {
		return fmt.Errorf("tipe tugas tidak valid: %s", req.TaskType)
	}

	if req.EvaluationType != "" && req.EvaluationType != "AUTO" && req.EvaluationType != "ADMIN_REVIEW" {
		return fmt.Errorf("tipe evaluasi tidak valid: %s (harus AUTO atau ADMIN_REVIEW)", req.EvaluationType)
	}

	if req.RewardCoins < 0 || req.RewardCoins > 1000000 {
		return fmt.Errorf("reward coins tidak valid (harus antara 1 dan 1,000,000)")
	}
	if req.RewardCoins == 0 {
		req.RewardCoins = 50
	}

	if req.RewardXP < 0 || req.RewardXP > 10000000 {
		return fmt.Errorf("reward xp tidak valid (harus antara 0 dan 10,000,000)")
	}
	if req.RewardXP == 0 {
		req.RewardXP = 100
	}

	if req.StepOrder <= 0 {
		req.StepOrder = 1
	}

	if req.Config == nil {
		req.Config = make(map[string]any)
	}

	// 1. Run capabilities inferred from task_type
	for _, capName := range caps {
		if validator, exists := e.validators[capName]; exists {
			if err := validator(req.Config); err != nil {
				return err
			}
		}
	}

	// 2. Task Composition: Detect composite capabilities embedded in Config
	// If config contains questions, validate quiz capability
	if req.Config["questions"] != nil && !contains(caps, "quiz") {
		if val, exists := e.validators["quiz"]; exists {
			if err := val(req.Config); err != nil {
				return err
			}
		}
	}

	// If config contains video/youtube_url, validate video capability
	if (req.Config["video_url"] != nil || req.Config["youtube_url"] != nil) && !contains(caps, "video") {
		if val, exists := e.validators["video"]; exists {
			if err := val(req.Config); err != nil {
				return err
			}
		}
	}

	// If config contains attachment_url/max_file_size_mb, validate document capability
	if (req.Config["attachment_url"] != nil || req.Config["max_file_size_mb"] != nil) && !contains(caps, "document") {
		if val, exists := e.validators["document"]; exists {
			if err := val(req.Config); err != nil {
				return err
			}
		}
	}

	// If config contains minimum_characters/maximum_characters, validate text capability
	if (req.Config["minimum_characters"] != nil || req.Config["maximum_characters"] != nil) && !contains(caps, "text") {
		if val, exists := e.validators["text"]; exists {
			if err := val(req.Config); err != nil {
				return err
			}
		}
	}

	// If config contains checklist items, validate checklist capability
	if req.Config["items"] != nil && !contains(caps, "checklist") {
		if val, exists := e.validators["checklist"]; exists {
			if err := val(req.Config); err != nil {
				return err
			}
		}
	}

	// If config contains photo capability hints, validate photo capability
	if (req.Config["max_files"] != nil || req.Config["accepted_mime_types"] != nil) && !contains(caps, "photo") {
		if val, exists := e.validators["photo"]; exists {
			if err := val(req.Config); err != nil {
				return err
			}
		}
	}

	// If config contains game capability hints, validate game capability
	if (req.Config["target_score"] != nil || req.Config["game"] != nil) && !contains(caps, "game") {
		if val, exists := e.validators["game"]; exists {
			if err := val(req.Config); err != nil {
				return err
			}
		}
	}

	return nil
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// ResolveEvaluationType determines whether a task evaluates as AUTO or ADMIN_REVIEW.
func ResolveEvaluationType(taskType string, explicitEvalType string) string {
	if explicitEvalType == "AUTO" || explicitEvalType == "ADMIN_REVIEW" {
		return explicitEvalType
	}
	switch strings.ToUpper(strings.TrimSpace(taskType)) {
	case "PHOTO_UPLOAD", "DOCUMENT_UPLOAD", "TEXT_RESPONSE", "PHOTO_PROOF":
		return "ADMIN_REVIEW"
	default:
		return "AUTO"
	}
}

// DefaultEngine is the global shared TaskEngine instance.
var DefaultEngine = NewEngine()
