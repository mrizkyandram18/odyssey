import type { TaskType } from '../../../shared/types'

/**
 * Single canonical builder for task-type → config mapping.
 *
 * Used by BOTH handleCreateTask and handleSaveEditTask in useAdminTasks.
 * Field shapes of both forms are structurally compatible, so one function
 * serves both. Caller-specific legacy differences are preserved explicitly
 * via options (documented below) — no behavioral cleanup.
 */

export interface TaskConfigQuestion {
  id: string
  question: string
  options: string[]
  correct_answer: string
}

export interface TaskConfigFields {
  task_type: TaskType | string
  video_url: string
  video_answer_mode: 'none' | 'quiz' | 'essay'
  video_mode: 'youtube' | 'recording'
  video_max_duration: number
  video_camera_facing: 'user' | 'environment'
  video_instruction: string
  questions: TaskConfigQuestion[]
  photo_min_count: number
  photo_camera_only: boolean
  photo_camera_facing: 'user' | 'environment'
  photo_camera_instruction: string
  doc_allowed_extensions: string
  doc_max_size_mb: number
  text_prompt: string
  text_min_chars: number
  text_max_chars: number
  game_type: string
  game_target_score: number
  game_max_moves: number
  game_scenario_json: string
}

export interface TaskConfigMessages {
  /** Create requires a YouTube URL; edit allows clearing it. */
  videoUrlRequired: string
  videoUrlInvalid: string
  /** Quiz-in-VIDEO wording (differs create vs edit, preserved per caller). */
  quizVideoEmpty: (index: number) => string
  quizVideoOptions: (index: number) => string
  quizVideoKey: (index: number) => string
  /** Standalone QUIZ wording (differs create vs edit, preserved per caller). */
  quizEmpty: (index: number) => string
  quizOptions: (index: number) => string
  quizKey: (index: number) => string
  /** Standalone TEXT_RESPONSE wording (differs create vs edit, preserved per caller). */
  textPromptEmpty: string
}

export interface BuildTaskConfigOptions {
  /**
   * Create requires a non-empty YouTube URL; edit allows empty (clear video)
   * but validates http(s) prefix when set. Preserved as-is.
   */
  requireVideoUrl: boolean
  /**
   * Edit preserves unknown custom keys on VIDEO-recording tasks and reuses
   * the stored config for GENERAL/legacy types. Create has no stored config.
   */
  preserveConfig?: Record<string, any> | null
  /**
   * Edit writes max_files alongside min_photos; create writes min_photos
   * only. Preserved as-is.
   */
  includePhotoMaxFiles: boolean
  /**
   * Edit carries a standalone QUIZ video_url over as youtube_url; create
   * drops it. Preserved as-is.
   */
  includeQuizYoutubeUrl: boolean
  messages: TaskConfigMessages
}

function fail(message: string): null {
  alert(message)
  return null
}

function firstInvalidQuestion(
  questions: TaskConfigQuestion[]
): { index: number; reason: 'question' | 'options' | 'key' } | null {
  for (let i = 0; i < questions.length; i++) {
    const q = questions[i]
    if (!q.question.trim()) return { index: i, reason: 'question' }
    if (q.options.some((opt) => !opt.trim())) return { index: i, reason: 'options' }
    if (!q.correct_answer.trim()) return { index: i, reason: 'key' }
  }
  return null
}

function parseDecisionScenario(rawJson: string): Record<string, any> | null {
  let parsed: any
  try {
    parsed = JSON.parse(rawJson)
  } catch {
    return fail('Scenario JSON tidak valid. Periksa format JSON.')
  }
  const scenario = parsed && typeof parsed === 'object' && parsed.scenario ? parsed.scenario : parsed
  if (!scenario || typeof scenario !== 'object' || !Array.isArray(scenario.events)) {
    return fail('Scenario harus memiliki object dengan key "events" (array).')
  }
  return scenario
}

function parseDocExtensions(text: string): string[] {
  return text
    .split(',')
    .map((s) => s.trim().toLowerCase().replace(/^\./, ''))
    .filter(Boolean)
}

function buildRecording(
  maxDuration: number,
  cameraFacing: string,
  instruction: string
): Record<string, any> {
  const recording: Record<string, any> = {
    enabled: true,
    max_duration_seconds: Math.max(1, Math.min(600, Number(maxDuration) || 60)),
    camera_facing: cameraFacing,
  }
  if (instruction.trim()) {
    recording.instruction = instruction.trim()
  }
  return recording
}

/**
 * Builds the task `config` payload for the given type + fields.
 * Returns null (after alerting, matching existing UX) when validation fails;
 * callers abort their submit on null.
 */
export function buildTaskConfig(
  fields: TaskConfigFields,
  options: BuildTaskConfigOptions
): Record<string, any> | null {
  const { messages } = options
  let config: Record<string, any> = {}

  if (fields.task_type === 'VIDEO') {
    if (fields.video_mode === 'recording') {
      const recording = buildRecording(
        fields.video_max_duration,
        fields.video_camera_facing,
        fields.video_instruction
      )
      config = { recording }
      // Edit-only: preserve unknown custom keys already on the task.
      if (options.preserveConfig && typeof options.preserveConfig === 'object') {
        for (const [k, v] of Object.entries(options.preserveConfig)) {
          if (k !== 'video_url' && k !== 'youtube_url' && k !== 'questions' && k !== 'prompt' && !(k in config)) {
            config[k] = v
          }
        }
      }
    } else {
      const url = fields.video_url.trim()
      if (options.requireVideoUrl) {
        if (!url) return fail(messages.videoUrlRequired)
      } else if (url && !url.startsWith('http')) {
        return fail(messages.videoUrlInvalid)
      }
      config = { video_url: url }
      if (url) config.youtube_url = url
      if (fields.video_answer_mode === 'quiz') {
        const bad = firstInvalidQuestion(fields.questions)
        if (bad) {
          if (bad.reason === 'question') return fail(messages.quizVideoEmpty(bad.index))
          if (bad.reason === 'options') return fail(messages.quizVideoOptions(bad.index))
          return fail(messages.quizVideoKey(bad.index))
        }
        config.questions = fields.questions
      } else if (fields.video_answer_mode === 'essay') {
        if (!fields.text_prompt.trim()) return fail('Prompt essay tidak boleh kosong')
        config.prompt = fields.text_prompt.trim()
        config.minimum_characters = fields.text_min_chars
        config.maximum_characters = fields.text_max_chars
      }
    }
  } else if (fields.task_type === 'QUIZ') {
    const bad = firstInvalidQuestion(fields.questions)
    if (bad) {
      if (bad.reason === 'question') return fail(messages.quizEmpty(bad.index))
      if (bad.reason === 'options') return fail(messages.quizOptions(bad.index))
      return fail(messages.quizKey(bad.index))
    }
    config = { questions: fields.questions }
    if (options.includeQuizYoutubeUrl && fields.video_url.trim()) {
      config.youtube_url = fields.video_url.trim()
    }
  } else if (fields.task_type === 'PHOTO_UPLOAD') {
    config = options.includePhotoMaxFiles
      ? {
          max_files: fields.photo_min_count,
          min_photos: fields.photo_min_count,
          camera_only: Boolean(fields.photo_camera_only),
        }
      : {
          min_photos: fields.photo_min_count,
          camera_only: Boolean(fields.photo_camera_only),
        }
    if (fields.photo_camera_only) {
      config.camera_facing = fields.photo_camera_facing
      if (fields.photo_camera_instruction.trim()) {
        config.camera_instruction = fields.photo_camera_instruction.trim()
      }
    }
  } else if (fields.task_type === 'DOCUMENT_UPLOAD') {
    config = {
      allowed_extensions: parseDocExtensions(fields.doc_allowed_extensions),
      max_file_size_mb: fields.doc_max_size_mb,
    }
  } else if (fields.task_type === 'TEXT_RESPONSE') {
    if (!fields.text_prompt.trim()) return fail(messages.textPromptEmpty)
    config = {
      prompt: fields.text_prompt.trim(),
      minimum_characters: fields.text_min_chars,
      maximum_characters: fields.text_max_chars,
    }
  } else if (fields.task_type === 'MINI_GAME') {
    config = {
      game: fields.game_type,
      target_score: fields.game_target_score,
      max_moves: fields.game_max_moves,
    }
    // Decision/finance scenario (optional): raw JSON parsed into config.scenario.
    // Server validates structure (ValidateTaskInput) and recomputes results.
    if (fields.game_scenario_json.trim()) {
      const scenario = parseDecisionScenario(fields.game_scenario_json)
      if (!scenario) return null
      config.scenario = scenario
    }
  } else {
    config = options.preserveConfig || {}
  }

  return config
}
