import React from 'react'
import { X } from 'lucide-react'

interface ImagePreviewModalProps {
  imageUrl: string | null
  onClose: () => void
}

export const ImagePreviewModal: React.FC<ImagePreviewModalProps> = ({ imageUrl, onClose }) => {
  if (!imageUrl) return null

  return (
    <div
      className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4"
      onClick={onClose}
    >
      <div
        className="relative max-w-4xl max-h-[90vh] bg-slate-900 border border-slate-700 rounded-2xl overflow-hidden shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <button
          onClick={onClose}
          className="absolute top-4 right-4 p-2 bg-slate-800/80 hover:bg-slate-700 text-slate-300 rounded-full z-10 transition-colors"
          title="Tutup Preview"
        >
          <X className="w-5 h-5" />
        </button>
        <img
          src={imageUrl}
          alt="Preview Bukti"
          className="w-full h-auto max-h-[85vh] object-contain rounded-2xl"
        />
      </div>
    </div>
  )
}
