import React, { useState, useEffect, useCallback } from 'react'
import { Sparkles, Plus, Check, AlertCircle, ToggleLeft, ToggleRight, Loader2, Award } from 'lucide-react'
import { adminCosmeticsApi } from '../../../../shared/lib/api'
import type { CosmeticCatalogItem } from '../../../../shared/types'
import { Avatar } from '../../../../shared/components/atoms/Avatar'

export const CosmeticsCatalogSection: React.FC = () => {
  const [items, setItems] = useState<CosmeticCatalogItem[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)
  const [successMsg, setSuccessMsg] = useState<string | null>(null)

  // Create modal state
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [isCreating, setIsCreating] = useState(false)
  const [modalError, setModalError] = useState<string | null>(null)
  const [newId, setNewId] = useState('')
  const [newName, setNewName] = useState('')
  const [newSlot, setNewSlot] = useState<'frame' | 'effect'>('frame')
  const [newAsset, setNewAsset] = useState('')
  const [newTier, setNewTier] = useState('1')
  const [newIsActive, setNewIsActive] = useState(true)

  const loadCosmetics = useCallback(async () => {
    setIsLoading(true)
    setErrorMsg(null)
    try {
      const res = await adminCosmeticsApi.getCosmetics()
      setItems(res.items || [])
    } catch (err: any) {
      setErrorMsg(err?.message || 'Gagal memuat katalog hadiah')
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    loadCosmetics()
  }, [loadCosmetics])

  const handleToggleActive = async (item: CosmeticCatalogItem) => {
    try {
      await adminCosmeticsApi.updateCosmetic(item.id, { is_active: !item.is_active })
      setItems((prev) =>
        prev.map((it) => (it.id === item.id ? { ...it, is_active: !it.is_active } : it))
      )
      setSuccessMsg(`Status ${item.name || item.id} berhasil diubah`)
      setTimeout(() => setSuccessMsg(null), 3000)
    } catch (err: any) {
      setErrorMsg(err?.message || 'Gagal mengubah status hadiah')
      setTimeout(() => setErrorMsg(null), 4000)
    }
  }

  const handleCreateCosmetic = async (e: React.FormEvent) => {
    e.preventDefault()
    setModalError(null)

    const trimmedId = newId.trim()
    const trimmedAsset = newAsset.trim()
    const tierNum = parseInt(newTier, 10)

    if (!trimmedId) {
      setModalError('ID hadiah wajib diisi')
      return
    }
    if (!trimmedAsset) {
      setModalError('Asset hadiah wajib diisi')
      return
    }
    if (isNaN(tierNum) || tierNum < 1 || tierNum > 3) {
      setModalError('Tier hadiah harus angka antara 1 sampai 3')
      return
    }

    setIsCreating(true)
    try {
      await adminCosmeticsApi.createCosmetic({
        id: trimmedId,
        name: newName.trim() || undefined,
        slot: newSlot,
        asset: trimmedAsset,
        tier: tierNum,
        is_active: newIsActive,
      })
      setIsModalOpen(false)
      setNewId('')
      setNewName('')
      setNewAsset('')
      setNewTier('1')
      setNewIsActive(true)
      setSuccessMsg(`Hadiah "${trimmedId}" berhasil ditambahkan!`)
      setTimeout(() => setSuccessMsg(null), 3000)
      loadCosmetics()
    } catch (err: any) {
      setModalError(err?.message || 'Gagal menambahkan hadiah')
    } finally {
      setIsCreating(false)
    }
  }

  return (
    <div className="p-4 sm:p-5 rounded-2xl bg-surface border border-border-subtle shadow-xs space-y-4">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2.5 pb-3 border-b border-border-subtle/60">
        <div>
          <h3 className="font-heading font-bold text-text-primary text-sm sm:text-base flex items-center gap-2">
            <Sparkles className="w-4 h-4 text-accent-gold" />
            <span>Katalog Hadiah & Koleksi</span>
          </h3>
          <p className="text-[11px] text-text-secondary mt-0.5">
            Koleksi visual bingkai avatar & efek animasi yang dapat diperoleh anggota saat membuka Kotak Hadiah.
          </p>
        </div>

        <button
          type="button"
          onClick={() => {
            setModalError(null)
            setIsModalOpen(true)
          }}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-accent-magic text-white text-xs font-bold hover:brightness-110 active:scale-[0.99] transition-all cursor-pointer self-start sm:self-auto"
        >
          <Plus className="w-3.5 h-3.5" />
          <span>Tambah Hadiah</span>
        </button>
      </div>

      {/* Notifications */}
      {errorMsg && (
        <div className="p-3 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-xs flex items-center gap-2">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{errorMsg}</span>
        </div>
      )}

      {successMsg && (
        <div className="p-3 rounded-xl bg-status-success/15 border border-status-success/30 text-status-success text-xs flex items-center gap-2">
          <Check className="w-4 h-4 shrink-0" />
          <span>{successMsg}</span>
        </div>
      )}

      {/* Cosmetics List */}
      {isLoading ? (
        <div className="p-8 text-center text-text-secondary text-xs flex items-center justify-center gap-2">
          <Loader2 className="w-4 h-4 animate-spin" />
          <span>Memuat katalog hadiah...</span>
        </div>
      ) : items.length === 0 ? (
        <div className="p-8 text-center text-text-secondary text-xs border border-dashed border-border-subtle rounded-xl space-y-2">
          <Award className="w-8 h-8 text-text-secondary/40 mx-auto" />
          <p className="font-bold text-text-primary text-sm">Belum Ada Hadiah Terdaftar</p>
          <p>Klik &quot;Tambah Hadiah&quot; untuk menambahkan bingkai atau efek visual baru.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
          {items.map((it) => {
            const stars = '★'.repeat(Math.max(1, Math.min(3, it.tier)))
            return (
              <div
                key={it.id}
                className={`p-4 rounded-2xl border transition-all flex flex-col justify-between gap-3 shadow-xs ${
                  it.is_active
                    ? 'bg-surface-elevated border-border-subtle hover:border-accent-magic/30'
                    : 'bg-surface/50 border-border-subtle/50 opacity-60'
                }`}
              >
                {/* Top Badge & Toggle */}
                <div className="flex items-center justify-between gap-2">
                  <span
                    className={`text-[9px] font-bold px-2 py-0.5 rounded-full uppercase tracking-wider ${
                      it.slot === 'frame'
                        ? 'bg-accent-magic/15 text-accent-magic border border-accent-magic/20'
                        : 'bg-accent-rare/15 text-accent-rare border border-accent-rare/20'
                    }`}
                  >
                    {it.slot === 'frame' ? 'Bingkai' : 'Efek'}
                  </span>

                  <button
                    type="button"
                    title={it.is_active ? 'Nonaktifkan hadiah' : 'Aktifkan hadiah'}
                    aria-label={it.is_active ? 'Nonaktifkan hadiah' : 'Aktifkan hadiah'}
                    onClick={() => handleToggleActive(it)}
                    className="cursor-pointer text-text-secondary hover:text-text-primary transition-colors shrink-0"
                  >
                    {it.is_active ? (
                      <ToggleRight className="w-6 h-6 text-status-success" />
                    ) : (
                      <ToggleLeft className="w-6 h-6 text-text-secondary/50" />
                    )}
                  </button>
                </div>

                {/* Central Live Avatar Preview */}
                <div className="py-2 flex flex-col items-center justify-center">
                  <Avatar
                    seed={it.id}
                    frame={it.slot === 'frame' ? it.asset : 'none'}
                    effect={it.slot === 'effect' ? it.asset : 'none'}
                    size="lg"
                  />
                </div>

                {/* Item Details */}
                <div className="text-center space-y-0.5">
                  <h4 className="text-xs font-bold text-text-primary line-clamp-1">
                    {it.name || it.id}
                  </h4>
                  <div className="flex items-center justify-center gap-1.5 text-[10px] text-text-secondary">
                    <span className="font-mono text-accent-gold">{stars}</span>
                    <span>•</span>
                    <span>Tier {it.tier}</span>
                  </div>
                </div>

                {/* Footer status */}
                <div className="pt-2 border-t border-border-subtle/60 text-center text-[10px]">
                  <span
                    className={`font-bold ${
                      it.is_active ? 'text-status-success' : 'text-text-secondary'
                    }`}
                  >
                    {it.is_active ? '● Aktif di Kotak Hadiah' : '○ Dinonaktifkan'}
                  </span>
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* Modal Tambah Hadiah */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-surface border border-border-subtle rounded-2xl w-full max-w-md p-5 space-y-4 shadow-xl">
            <div className="flex items-center justify-between pb-2 border-b border-border-subtle">
              <h4 className="font-bold text-text-primary text-sm flex items-center gap-2">
                <Plus className="w-4 h-4 text-accent-magic" />
                <span>Tambah Hadiah Baru</span>
              </h4>
              <button
                type="button"
                onClick={() => setIsModalOpen(false)}
                className="text-text-secondary hover:text-text-primary text-xs font-bold cursor-pointer"
              >
                ✕
              </button>
            </div>

            {modalError && (
              <div className="p-2.5 rounded-xl bg-status-error/15 border border-status-error/30 text-status-error text-xs flex items-center gap-2">
                <AlertCircle className="w-4 h-4 shrink-0" />
                <span>{modalError}</span>
              </div>
            )}

            {/* Live Visual Preview of New Reward */}
            <div className="p-3 rounded-xl bg-surface-elevated border border-border-subtle flex items-center gap-3">
              <Avatar
                seed="preview-new"
                frame={newSlot === 'frame' ? newAsset || 'none' : 'none'}
                effect={newSlot === 'effect' ? newAsset || 'none' : 'none'}
                size="md"
              />
              <div className="space-y-0.5 min-w-0">
                <p className="text-xs font-bold text-text-primary truncate">
                  {newName.trim() || 'Pratinjau Hadiah'}
                </p>
                <p className="text-[11px] text-text-secondary">
                  Tipe: {newSlot === 'frame' ? 'Bingkai Avatar' : 'Efek Animasi'} • Tier {newTier}
                </p>
              </div>
            </div>

            <form onSubmit={handleCreateCosmetic} className="space-y-3">
              <div className="space-y-1">
                <label className="text-xs font-bold text-text-secondary">
                  ID Hadiah <span className="text-status-error">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder="contoh: frame-emerald atau effect-fire"
                  value={newId}
                  onChange={(e) => setNewId(e.target.value)}
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-mono text-text-primary focus:outline-none focus:border-accent-magic"
                />
              </div>

              <div className="space-y-1">
                <label className="text-xs font-bold text-text-secondary">
                  Nama Hadiah (Display Name)
                </label>
                <input
                  type="text"
                  placeholder="contoh: Bingkai Zamrud"
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs text-text-primary focus:outline-none focus:border-accent-magic"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1">
                  <label className="text-xs font-bold text-text-secondary">
                    Jenis Hadiah (Slot) <span className="text-status-error">*</span>
                  </label>
                  <select
                    value={newSlot}
                    onChange={(e) => setNewSlot(e.target.value as 'frame' | 'effect')}
                    className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-bold text-text-primary focus:outline-none focus:border-accent-magic cursor-pointer"
                  >
                    <option value="frame">Bingkai Avatar (frame)</option>
                    <option value="effect">Efek Visual (effect)</option>
                  </select>
                </div>

                <div className="space-y-1">
                  <label className="text-xs font-bold text-text-secondary">
                    Kode Asset Visual <span className="text-status-error">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="contoh: emerald / sparkle"
                    value={newAsset}
                    onChange={(e) => setNewAsset(e.target.value)}
                    className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-mono text-text-primary focus:outline-none focus:border-accent-magic"
                  />
                </div>
              </div>

              {/* Quick asset suggestion chips */}
              <div className="space-y-1">
                <span className="text-[10px] text-text-secondary">Saran asset yang sudah didukung visual:</span>
                <div className="flex flex-wrap gap-1">
                  {(newSlot === 'frame' ? ['gold'] : ['sparkle', 'float', 'trail']).map((sugg) => (
                    <button
                      key={sugg}
                      type="button"
                      onClick={() => {
                        setNewAsset(sugg)
                        if (!newId) setNewId(`${newSlot}-${sugg}`)
                        if (!newName) setNewName(sugg === 'gold' ? 'Bingkai Emas' : `Efek ${sugg.charAt(0).toUpperCase() + sugg.slice(1)}`)
                      }}
                      className="px-2 py-0.5 rounded-md bg-surface-elevated border border-border-subtle text-[10px] text-text-primary hover:border-accent-magic cursor-pointer transition-colors"
                    >
                      {sugg}
                    </button>
                  ))}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1">
                  <label className="text-xs font-bold text-text-secondary">
                    Tingkat Hadiah Awal (Tier 1–3) <span className="text-status-error">*</span>
                  </label>
                  <input
                    type="number"
                    min={1}
                    max={3}
                    required
                    value={newTier}
                    onChange={(e) => setNewTier(e.target.value)}
                    className="w-full p-2.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs font-mono text-text-primary focus:outline-none focus:border-accent-magic"
                  />
                </div>

                <div className="flex items-center gap-2 pt-5">
                  <input
                    type="checkbox"
                    id="cb-is-active"
                    checked={newIsActive}
                    onChange={(e) => setNewIsActive(e.target.checked)}
                    className="w-4 h-4 rounded text-accent-magic focus:ring-accent-magic cursor-pointer"
                  />
                  <label htmlFor="cb-is-active" className="text-xs font-bold text-text-primary cursor-pointer">
                    Aktif di Kotak Hadiah
                  </label>
                </div>
              </div>

              <div className="flex items-center justify-end gap-2 pt-3 border-t border-border-subtle">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-3 py-2 rounded-xl text-xs font-bold text-text-secondary hover:bg-surface-elevated transition-colors cursor-pointer"
                >
                  Batal
                </button>
                <button
                  type="submit"
                  disabled={isCreating}
                  className="px-4 py-2 rounded-xl bg-accent-magic text-white text-xs font-bold hover:brightness-110 active:scale-[0.99] disabled:opacity-50 transition-all cursor-pointer"
                >
                  {isCreating ? 'Menyimpan...' : 'Simpan Hadiah'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
