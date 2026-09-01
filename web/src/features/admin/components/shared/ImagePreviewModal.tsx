import React from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { X } from 'lucide-react'

interface ImagePreviewModalProps {
  imageUrl: string | null
  onClose: () => void
}

export const ImagePreviewModal: React.FC<ImagePreviewModalProps> = ({ imageUrl, onClose }) => {
  return (
    <AnimatePresence>
      {imageUrl && (
        <div
          className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4"
          onClick={onClose}
        >
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
            transition={{ duration: 0.15 }}
            className="relative max-w-4xl max-h-[90vh] bg-surface-elevated border border-border-subtle rounded-2xl overflow-hidden shadow-2xl"
            onClick={(e) => e.stopPropagation()}
          >
            <button
              type="button"
              onClick={onClose}
              className="absolute top-3 right-3 p-2 bg-black/50 hover:bg-black/70 text-white rounded-full z-10 transition-colors"
              title="Tutup Preview"
              aria-label="Tutup Preview"
            >
              <X className="w-5 h-5" />
            </button>
            <img
              src={imageUrl}
              alt="Preview Bukti"
              className="w-full h-auto max-h-[85vh] object-contain rounded-2xl"
            />
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  )
}
