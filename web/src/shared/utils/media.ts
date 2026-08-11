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
