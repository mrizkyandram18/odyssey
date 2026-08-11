/** Helpers for Photo (and future media) creative submissions. */

export interface PhotoPayload {
  v?: number
  photo: string
  caption?: string
}

export const MAX_PHOTO_FILE_BYTES = 5 * 1024 * 1024
export const MAX_PHOTO_CAPTION = 280

export function parsePhotoPayload(content: string): PhotoPayload | null {
  if (!content || !content.trim().startsWith('{')) return null
  try {
    const parsed = JSON.parse(content) as PhotoPayload
    if (typeof parsed !== 'object' || parsed === null || typeof parsed.photo !== 'string') {
      return null
    }
    return {
      v: parsed.v,
      photo: parsed.photo,
      caption: typeof parsed.caption === 'string' ? parsed.caption : undefined,
    }
  } catch {
    return null
  }
}

export function buildPhotoPayload(photo: string, caption = ''): string {
  const payload: PhotoPayload = {
    v: 1,
    photo: photo.trim(),
    caption: caption.trim(),
  }
  return JSON.stringify(payload)
}

export function isImageDataURL(uri: string): boolean {
  if (!uri || typeof uri !== 'string') return false
  if (!uri.startsWith('data:image/')) return false
  return /;base64,/.test(uri)
}

export function isImageFile(file: File): boolean {
  return typeof file?.type === 'string' && file.type.startsWith('image/')
}

export function dataURLToBlob(dataURL: string): Blob {
  const match = /^data:(image\/[a-z+-.]+);base64,/i.exec(dataURL)
  if (!match) {
    throw new Error('not a valid base64 image data URL')
  }
  const mime = match[1]
  const bstr = atob(dataURL.slice(match[0].length))
  const u8 = new Uint8Array(bstr.length)
  for (let i = 0; i < bstr.length; i++) {
    u8[i] = bstr.charCodeAt(i)
  }
  return new Blob([u8], { type: mime })
}

export function fileToDataURL(file: File): Promise<string> {
  return new Promise<string>((resolve, reject) => {
    if (!isImageFile(file)) {
      reject(new Error('only image files are allowed'))
      return
    }
    if (file.size > MAX_PHOTO_FILE_BYTES) {
      reject(new Error('photo exceeds maximum size'))
      return
    }
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = () => reject(reader.error ?? new Error('failed to read file'))
    reader.readAsDataURL(file)
  })
}

/** Video submission helpers, mirroring the PHOTO contract. */

export interface VideoPayload {
  v?: number
  video: string
  caption?: string
}

export const MAX_VIDEO_FILE_BYTES = 5 * 1024 * 1024
export const MAX_VIDEO_CAPTION = 280

export function parseVideoPayload(content: string): VideoPayload | null {
  if (!content || !content.trim().startsWith('{')) return null
  try {
    const parsed = JSON.parse(content) as VideoPayload
    if (typeof parsed !== 'object' || parsed === null || typeof parsed.video !== 'string') {
      return null
    }
    return {
      v: parsed.v,
      video: parsed.video,
      caption: typeof parsed.caption === 'string' ? parsed.caption : undefined,
    }
  } catch {
    return null
  }
}

export function buildVideoPayload(video: string, caption = ''): string {
  const payload: VideoPayload = {
    v: 1,
    video: video.trim(),
    caption: caption.trim(),
  }
  return JSON.stringify(payload)
}

export function isVideoDataURL(uri: string): boolean {
  if (!uri || typeof uri !== 'string') return false
  if (!uri.startsWith('data:video/')) return false
  return /;base64,/.test(uri)
}

export function isVideoFile(file: File): boolean {
  return typeof file?.type === 'string' && file.type.startsWith('video/')
}

export function isVideoContainerSignature(mime: string, bytes: Uint8Array): boolean {
  if (bytes.length < 4) return false
  switch (mime) {
    case 'video/mp4':
      return bytes.length >= 8 && String.fromCharCode(bytes[4], bytes[5], bytes[6], bytes[7]) === 'ftyp'
    case 'video/webm':
      return bytes[0] === 0x1a && bytes[1] === 0x45 && bytes[2] === 0xdf && bytes[3] === 0xa3
    case 'video/ogg':
      return String.fromCharCode(bytes[0], bytes[1], bytes[2], bytes[3]) === 'OggS'
  }
  return false
}

export function fileToVideoDataURL(file: File): Promise<string> {
  return new Promise<string>((resolve, reject) => {
    if (!isVideoFile(file)) {
      reject(new Error('only video files are allowed'))
      return
    }
    if (file.size > MAX_VIDEO_FILE_BYTES) {
      reject(new Error('video exceeds maximum size'))
      return
    }
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = () => reject(reader.error ?? new Error('failed to read file'))
    reader.readAsDataURL(file)
  })
}
