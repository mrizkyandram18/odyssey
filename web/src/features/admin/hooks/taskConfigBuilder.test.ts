// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  buildTaskConfig,
  type BuildTaskConfigOptions,
  type TaskConfigFields,
} from './taskConfigBuilder'

const baseFields: TaskConfigFields = {
  task_type: 'VIDEO',
  video_url: '',
  video_answer_mode: 'none',
  video_mode: 'youtube',
  video_max_duration: 60,
  video_camera_facing: 'user',
  video_instruction: '',
  questions: [{ id: '1', question: 'Q?', options: ['A', 'B'], correct_answer: 'A' }],
  photo_min_count: 1,
  photo_camera_only: false,
  photo_camera_facing: 'environment',
  photo_camera_instruction: '',
  doc_allowed_extensions: 'pdf,docx,xlsx,txt',
  doc_max_size_mb: 10,
  text_prompt: '',
  text_min_chars: 10,
  text_max_chars: 500,
  game_type: 'MEMORY_MATCH',
  game_target_score: 500,
  game_max_moves: 20,
  game_scenario_json: '',
}

// Option sets mirroring the two historical call sites (create vs edit parity).
const createOptions: BuildTaskConfigOptions = {
  requireVideoUrl: true,
  includePhotoMaxFiles: false,
  includeQuizYoutubeUrl: false,
  messages: {
    videoUrlRequired: 'URL Video YouTube wajib diisi',
    videoUrlInvalid: 'URL video harus http(s)',
    quizVideoEmpty: (i) => `Pertanyaan ke-${i + 1} belum diisi`,
    quizVideoOptions: (i) => `Opsi ke-${i + 1} belum lengkap`,
    quizVideoKey: (i) => `Kunci #${i + 1} wajib dipilih`,
    quizEmpty: (i) => `Pertanyaan ke-${i + 1} belum diisi`,
    quizOptions: (i) => `Semua opsi pilihan jawaban pada pertanyaan ke-${i + 1} wajib diisi`,
    quizKey: (i) => `Kunci jawaban pada pertanyaan ke-${i + 1} wajib dipilih`,
    textPromptEmpty: 'Instruksi/pertanyaan esai tidak boleh kosong',
  },
}

const editOptions: BuildTaskConfigOptions = {
  requireVideoUrl: false,
  preserveConfig: { custom_flag: true, video_url: 'old' },
  includePhotoMaxFiles: true,
  includeQuizYoutubeUrl: true,
  messages: {
    videoUrlRequired: 'URL Video YouTube wajib diisi',
    videoUrlInvalid: 'URL video harus http(s)',
    quizVideoEmpty: (i) => `Pertanyaan #${i + 1} belum diisi`,
    quizVideoOptions: (i) => `Opsi #${i + 1} belum lengkap`,
    quizVideoKey: (i) => `Kunci #${i + 1} belum dipilih`,
    quizEmpty: (i) => `Pertanyaan #${i + 1} belum diisi`,
    quizOptions: (i) => `Opsi pertanyaan #${i + 1} belum lengkap`,
    quizKey: (i) => `Kunci jawaban #${i + 1} belum dipilih`,
    textPromptEmpty: 'Prompt esai tidak boleh kosong',
  },
}

let alertMock: ReturnType<typeof vi.fn>

beforeEach(() => {
  alertMock = vi.fn()
  vi.stubGlobal('alert', alertMock)
})

describe('buildTaskConfig — VIDEO', () => {
  it('create requires a YouTube URL', () => {
    const out = buildTaskConfig({ ...baseFields, task_type: 'VIDEO', video_url: '  ' }, createOptions)
    expect(out).toBeNull()
    expect(alertMock).toHaveBeenCalledWith('URL Video YouTube wajib diisi')
  })

  it('create sets both video_url and youtube_url', () => {
    const out = buildTaskConfig(
      { ...baseFields, task_type: 'VIDEO', video_url: 'https://www.youtube.com/watch?v=x' },
      createOptions
    )
    expect(out).toEqual({
      video_url: 'https://www.youtube.com/watch?v=x',
      youtube_url: 'https://www.youtube.com/watch?v=x',
    })
  })

  it('edit allows empty URL and rejects non-http(s)', () => {
    expect(
      buildTaskConfig({ ...baseFields, task_type: 'VIDEO', video_url: '' }, editOptions)
    ).toEqual({ video_url: '' })
    expect(
      buildTaskConfig({ ...baseFields, task_type: 'VIDEO', video_url: 'ftp://x' }, editOptions)
    ).toBeNull()
    expect(alertMock).toHaveBeenCalledWith('URL video harus http(s)')
  })

  it('builds recording config with clamped duration', () => {
    const out = buildTaskConfig(
      {
        ...baseFields,
        task_type: 'VIDEO',
        video_mode: 'recording',
        video_max_duration: 9999,
        video_camera_facing: 'environment',
        video_instruction: 'Intro 60 detik',
      },
      createOptions
    )
    expect(out).toEqual({
      recording: {
        enabled: true,
        max_duration_seconds: 600,
        camera_facing: 'environment',
        instruction: 'Intro 60 detik',
      },
    })
  })

  it('edit preserves unknown custom keys on recording tasks', () => {
    const out = buildTaskConfig(
      { ...baseFields, task_type: 'VIDEO', video_mode: 'recording' },
      editOptions
    )
    expect(out).toEqual({
      recording: { enabled: true, max_duration_seconds: 60, camera_facing: 'user' },
      custom_flag: true,
    })
  })

  it('embeds quiz questions with caller wording on invalid input', () => {
    const bad = {
      ...baseFields,
      task_type: 'VIDEO',
      video_url: 'https://x',
      video_answer_mode: 'quiz' as const,
      questions: [{ id: '1', question: '', options: ['A', 'B'], correct_answer: 'A' }],
    }
    expect(buildTaskConfig(bad, createOptions)).toBeNull()
    expect(alertMock).toHaveBeenCalledWith('Pertanyaan ke-1 belum diisi')
    alertMock.mockClear()
    expect(buildTaskConfig(bad, editOptions)).toBeNull()
    expect(alertMock).toHaveBeenCalledWith('Pertanyaan #1 belum diisi')
  })

  it('embeds essay prompt with min/max chars', () => {
    const out = buildTaskConfig(
      {
        ...baseFields,
        task_type: 'VIDEO',
        video_url: 'https://x',
        video_answer_mode: 'essay' as const,
        text_prompt: 'Jelaskan!',
        text_min_chars: 80,
        text_max_chars: 1000,
      },
      createOptions
    )
    expect(out).toEqual({
      video_url: 'https://x',
      youtube_url: 'https://x',
      prompt: 'Jelaskan!',
      minimum_characters: 80,
      maximum_characters: 1000,
    })
  })
})

describe('buildTaskConfig — QUIZ', () => {
  it('create attaches questions and drops video_url', () => {
    const out = buildTaskConfig(
      { ...baseFields, task_type: 'QUIZ', video_url: 'https://x' },
      createOptions
    )
    expect(out).toEqual({ questions: baseFields.questions })
  })

  it('edit carries youtube_url over', () => {
    const out = buildTaskConfig(
      { ...baseFields, task_type: 'QUIZ', video_url: 'https://x' },
      editOptions
    )
    expect(out).toEqual({ questions: baseFields.questions, youtube_url: 'https://x' })
  })

  it('rejects incomplete options with caller wording', () => {
    const bad = {
      ...baseFields,
      task_type: 'QUIZ',
      questions: [{ id: '1', question: 'Q?', options: ['A', ''], correct_answer: 'A' }],
    }
    expect(buildTaskConfig(bad, createOptions)).toBeNull()
    expect(alertMock).toHaveBeenCalledWith('Semua opsi pilihan jawaban pada pertanyaan ke-1 wajib diisi')
    alertMock.mockClear()
    expect(buildTaskConfig(bad, editOptions)).toBeNull()
    expect(alertMock).toHaveBeenCalledWith('Opsi pertanyaan #1 belum lengkap')
  })
})

describe('buildTaskConfig — PHOTO_UPLOAD', () => {
  it('create writes min_photos without max_files', () => {
    expect(
      buildTaskConfig({ ...baseFields, task_type: 'PHOTO_UPLOAD' }, createOptions)
    ).toEqual({ min_photos: 1, camera_only: false })
  })

  it('edit writes max_files alongside min_photos with camera options', () => {
    expect(
      buildTaskConfig(
        {
          ...baseFields,
          task_type: 'PHOTO_UPLOAD',
          photo_min_count: 2,
          photo_camera_only: true,
          photo_camera_facing: 'user',
          photo_camera_instruction: 'Ambil foto',
        },
        editOptions
      )
    ).toEqual({
      max_files: 2,
      min_photos: 2,
      camera_only: true,
      camera_facing: 'user',
      camera_instruction: 'Ambil foto',
    })
  })
})

describe('buildTaskConfig — DOCUMENT_UPLOAD', () => {
  it('normalizes extensions', () => {
    expect(
      buildTaskConfig(
        { ...baseFields, task_type: 'DOCUMENT_UPLOAD', doc_allowed_extensions: '.PDF, docx ,,txt' },
        createOptions
      )
    ).toEqual({ allowed_extensions: ['pdf', 'docx', 'txt'], max_file_size_mb: 10 })
  })
})

describe('buildTaskConfig — TEXT_RESPONSE', () => {
  it('rejects blank prompt with caller wording', () => {
    expect(
      buildTaskConfig({ ...baseFields, task_type: 'TEXT_RESPONSE', text_prompt: '  ' }, createOptions)
    ).toBeNull()
    expect(alertMock).toHaveBeenCalledWith('Instruksi/pertanyaan esai tidak boleh kosong')
    alertMock.mockClear()
    expect(
      buildTaskConfig({ ...baseFields, task_type: 'TEXT_RESPONSE', text_prompt: '' }, editOptions)
    ).toBeNull()
    expect(alertMock).toHaveBeenCalledWith('Prompt esai tidak boleh kosong')
  })

  it('builds prompt shape', () => {
    expect(
      buildTaskConfig(
        { ...baseFields, task_type: 'TEXT_RESPONSE', text_prompt: 'Refleksi?', text_min_chars: 80, text_max_chars: 1000 },
        createOptions
      )
    ).toEqual({ prompt: 'Refleksi?', minimum_characters: 80, maximum_characters: 1000 })
  })
})

describe('buildTaskConfig — MINI_GAME', () => {
  it('builds base game config without scenario', () => {
    expect(
      buildTaskConfig({ ...baseFields, task_type: 'MINI_GAME' }, createOptions)
    ).toEqual({ game: 'MEMORY_MATCH', target_score: 500, max_moves: 20 })
  })

  it('unwraps { scenario } and accepts bare scenario objects', () => {
    const wrapped = JSON.stringify({ scenario: { initial_balance: 0, events: [{ id: 'e1' }] } })
    const bare = JSON.stringify({ initial_balance: 0, events: [{ id: 'e1' }] })
    for (const raw of [wrapped, bare]) {
      const out = buildTaskConfig(
        { ...baseFields, task_type: 'MINI_GAME', game_scenario_json: raw },
        createOptions
      )
      expect(out?.scenario).toEqual({ initial_balance: 0, events: [{ id: 'e1' }] })
    }
  })

  it('rejects invalid JSON and non-event scenarios', () => {
    expect(
      buildTaskConfig({ ...baseFields, task_type: 'MINI_GAME', game_scenario_json: '{oops' }, createOptions)
    ).toBeNull()
    expect(alertMock).toHaveBeenCalledWith('Scenario JSON tidak valid. Periksa format JSON.')
    alertMock.mockClear()
    expect(
      buildTaskConfig(
        { ...baseFields, task_type: 'MINI_GAME', game_scenario_json: '{"foo":1}' },
        createOptions
      )
    ).toBeNull()
    expect(alertMock).toHaveBeenCalledWith('Scenario harus memiliki object dengan key "events" (array).')
  })
})

describe('buildTaskConfig — GENERAL/legacy passthrough', () => {
  it('reuses stored config on edit, empty object on create', () => {
    expect(buildTaskConfig({ ...baseFields, task_type: 'GENERAL' }, createOptions)).toEqual({})
    expect(buildTaskConfig({ ...baseFields, task_type: 'GENERAL' }, editOptions)).toEqual({
      custom_flag: true,
      video_url: 'old',
    })
  })
})
