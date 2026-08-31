/**
 * Client-side image compressor & watermark tool for mobile camera uploads.
 * Ensures pictures taken on modern phones (5-15 MB) are scaled down to <= 1280px,
 * compressed as JPEG/WebP (~150-300 KB), and tagged with a clean timestamp watermark.
 */

export interface CompressOptions {
  maxWidth?: number
  maxHeight?: number
  quality?: number
  watermarkText?: string
}

export async function compressImage(
  file: File,
  options: CompressOptions = {}
): Promise<{ file: File; dataUrl: string; width: number; height: number }> {
  const maxWidth = options.maxWidth || 1280
  const maxHeight = options.maxHeight || 1280
  const quality = options.quality || 0.7

  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onerror = reject
    reader.onload = () => {
      const img = new Image()
      img.onerror = reject
      img.onload = () => {
        let width = img.width
        let height = img.height

        // Calculate aspect ratio
        if (width > height) {
          if (width > maxWidth) {
            height = Math.round((height * maxWidth) / width)
            width = maxWidth
          }
        } else {
          if (height > maxHeight) {
            width = Math.round((width * maxHeight) / height)
            height = maxHeight
          }
        }

        const canvas = document.createElement('canvas')
        canvas.width = width
        canvas.height = height
        const ctx = canvas.getContext('2d')
        if (!ctx) {
          return reject(new Error('Canvas context unavailable'))
        }

        // Draw compressed image
        ctx.drawImage(img, 0, 0, width, height)

        // Draw timestamp watermark in bottom right corner
        const timestamp = options.watermarkText || new Date().toLocaleString('id-ID', {
          dateStyle: 'medium',
          timeStyle: 'medium',
        })

        const fontSize = Math.max(14, Math.round(width * 0.025))
        ctx.font = `600 ${fontSize}px sans-serif`
        const textWidth = ctx.measureText(timestamp).width

        // Background pill for contrast
        const padding = 8
        const bgX = width - textWidth - padding * 2 - 12
        const bgY = height - fontSize - padding * 2 - 12
        ctx.fillStyle = 'rgba(0, 0, 0, 0.65)'
        ctx.beginPath()
        ctx.roundRect(bgX, bgY, textWidth + padding * 2, fontSize + padding * 2, 6)
        ctx.fill()

        // Watermark text
        ctx.fillStyle = '#FFFFFF'
        ctx.textBaseline = 'middle'
        ctx.fillText(timestamp, bgX + padding, bgY + padding + fontSize / 2)

        // Export to Blob
        canvas.toBlob(
          (blob) => {
            if (!blob) {
              return reject(new Error('Failed to generate compressed blob'))
            }
            const cleanName = file.name.replace(/\.[^/.]+$/, '') + '.jpg'
            const compressedFile = new File([blob], cleanName, {
              type: 'image/jpeg',
              lastModified: Date.now(),
            })
            const dataUrl = canvas.toDataURL('image/jpeg', quality)
            resolve({
              file: compressedFile,
              dataUrl,
              width,
              height,
            })
          },
          'image/jpeg',
          quality
        )
      }
      img.src = reader.result as string
    }
    reader.readAsDataURL(file)
  })
}

export async function uploadTaskProof(
  file: File
): Promise<{ file_url: string; file_name: string; file_size: number }> {
  const formData = new FormData()
  formData.append('file', file)

  const token = localStorage.getItem('odyssey_session_token') || ''
  const response = await fetch('/api/tasks/upload', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'X-User-Session': token,
    },
    body: formData,
  })

  if (!response.ok) {
    const errText = await response.text()
    throw new Error('Gagal mengunggah file bukti: ' + errText)
  }

  return response.json()
}
