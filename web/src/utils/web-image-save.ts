export type SaveImageResult = 'shared' | 'preview' | 'downloaded' | 'noop'

export interface SaveImageByUrlOptions {
  title?: string
  blobUrl?: string
}

function sanitizeFileNamePart(value: string | number | null | undefined) {
  return String(value ?? '')
    .trim()
    .replace(/[\\/:*?"<>|]+/g, '_')
    .replace(/\s+/g, ' ')
}

export function isAppleSafariBrowser() {
  if (typeof navigator === 'undefined') return false
  const userAgent = navigator.userAgent || ''
  const vendor = navigator.vendor || ''
  const isAppleMobile = /iP(hone|od|ad)/i.test(userAgent)
  const isSafari = /Safari/i.test(userAgent) && !/CriOS|FxiOS|EdgiOS|OPiOS|Android/i.test(userAgent) && /Apple/i.test(vendor)
  return isAppleMobile && isSafari
}

export function buildReceiptFileName(orderNo: string) {
  const safeOrderNo = sanitizeFileNamePart(orderNo) || 'receipt'
  return `小票_${safeOrderNo}.png`
}

export function buildOrderCareReportFileName(orderNo: string, petName: string) {
  const safeOrderNo = sanitizeFileNamePart(orderNo) || 'NO'
  const safePetName = sanitizeFileNamePart(petName) || '猫咪'
  return `护理报告_${safeOrderNo}_${safePetName}.png`
}

export function openImagePreviewWindow(src: string, title = '图片预览') {
  if (!src || typeof window === 'undefined') return false
  const previewWindow = window.open('', '_blank')
  if (!previewWindow) return false

  const escapedTitle = title
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')

  const doc = previewWindow.document
  doc.open()
  doc.write(`<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1,viewport-fit=cover"><title>${escapedTitle}</title><style>html,body{margin:0;padding:0;background:#111;}body{min-height:100vh;display:flex;align-items:flex-start;justify-content:center;padding:16px;box-sizing:border-box;}img{display:block;max-width:100%;height:auto;border-radius:12px;-webkit-user-select:auto;user-select:auto;-webkit-touch-callout:default;box-shadow:0 6px 24px rgba(0,0,0,.24);}</style></head><body></body></html>`)
  doc.close()

  const img = doc.createElement('img')
  img.src = src
  img.alt = title
  doc.body.appendChild(img)
  return true
}

export function dataUrlToBlob(dataUrl: string): Blob {
  const parts = dataUrl.split(',')
  const mime = parts[0]?.match(/:(.*?);/)?.[1] || 'image/png'
  const raw = atob(parts[1] || '')
  const buffer = new Uint8Array(raw.length)
  for (let index = 0; index < raw.length; index += 1) {
    buffer[index] = raw.charCodeAt(index)
  }
  return new Blob([buffer], { type: mime })
}

async function resolveImageBlob(src: string) {
  if (!src || typeof fetch !== 'function') return null
  if (src.startsWith('data:')) {
    return dataUrlToBlob(src)
  }
  try {
    const res = await fetch(src)
    if (!res.ok) return null
    return await res.blob()
  } catch {
    return null
  }
}

export async function saveImageByUrl(src: string, fileName: string, options: SaveImageByUrlOptions = {}): Promise<SaveImageResult> {
  if (!src || typeof window === 'undefined' || typeof document === 'undefined') {
    return 'noop'
  }

  if (isAppleSafariBrowser() && openImagePreviewWindow(src, options.title || fileName)) {
    return 'preview'
  }

  const blob = await resolveImageBlob(src)

  if (blob && typeof navigator !== 'undefined' && typeof navigator.share === 'function' && typeof File === 'function') {
    const file = new File([blob], fileName, { type: blob.type || 'image/png' })
    const canShareFiles = typeof navigator.canShare !== 'function' || navigator.canShare({ files: [file] })
    if (canShareFiles) {
      try {
        await navigator.share({
          title: options.title || fileName,
          files: [file],
        })
        return 'shared'
      } catch (error) {
        if (error instanceof DOMException && error.name === 'AbortError') {
          return 'noop'
        }
      }
    }
  }

  const objectUrl = options.blobUrl || (blob ? URL.createObjectURL(blob) : '')
  const anchor = document.createElement('a')
  anchor.href = objectUrl || src
  anchor.download = fileName
  anchor.rel = 'noopener'
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)

  if (!options.blobUrl && objectUrl) {
    setTimeout(() => URL.revokeObjectURL(objectUrl), 0)
  }

  return 'downloaded'
}
