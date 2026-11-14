import { useState } from 'react'
import type { ChangeEvent, DragEvent } from 'react'
import { FileText, Upload, X } from 'lucide-react'
import { clsx } from 'clsx'
import styles from './FileUpload.module.css'

type FileUploadProps = {
  label: string
  files: File[]
  onChange: (files: File[]) => void
  hint?: string
  accept?: string
  maxSizeMb?: number
  error?: string
}

function formatSize(bytes: number) {
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} КБ`
  return `${(bytes / 1024 / 1024).toFixed(1).replace('.', ',')} МБ`
}

export function FileUpload({ label, files, onChange, hint, accept, maxSizeMb = 10, error }: FileUploadProps) {
  const [dragging, setDragging] = useState(false)
  const [localError, setLocalError] = useState('')

  function addFiles(list: FileList | null) {
    if (!list) return
    const allowed = accept ? accept.split(',').map((ext) => ext.trim().toLowerCase()) : null
    const added: File[] = []
    let problem = ''

    for (const file of Array.from(list)) {
      const ext = '.' + file.name.split('.').pop()?.toLowerCase()
      if (allowed && !allowed.includes(ext)) {
        problem = `Файл «${file.name}» не подходит: можно ${allowed.join(', ')}`
      } else if (file.size > maxSizeMb * 1024 * 1024) {
        problem = `Файл «${file.name}» больше ${maxSizeMb} МБ`
      } else {
        added.push(file)
      }
    }

    setLocalError(problem)
    if (added.length > 0) onChange([...files, ...added])
  }

  function handleInput(event: ChangeEvent<HTMLInputElement>) {
    addFiles(event.target.files)
    event.target.value = ''
  }

  function handleDrop(event: DragEvent<HTMLLabelElement>) {
    event.preventDefault()
    setDragging(false)
    addFiles(event.dataTransfer.files)
  }

  function remove(index: number) {
    onChange(files.filter((_, i) => i !== index))
  }

  const shownError = localError || error

  return (
    <div className={styles.field}>
      <span className={styles.label}>{label}</span>

      <label
        className={clsx(styles.drop, dragging ? styles.dragging : '')}
        onDragOver={(event) => {
          event.preventDefault()
          setDragging(true)
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={handleDrop}
      >
        <input type="file" multiple accept={accept} className={styles.input} onChange={handleInput} />
        <Upload size={20} strokeWidth={1.75} aria-hidden="true" />
        <span>
          Перетащите файлы сюда или <strong>выберите на компьютере</strong>
        </span>
      </label>

      {files.length > 0 ? (
        <ul className={styles.list}>
          {files.map((file, index) => (
            <li key={`${file.name}-${index}`} className={styles.file}>
              <FileText size={16} className={styles.fileIcon} aria-hidden="true" />
              <span className={styles.fileName}>{file.name}</span>
              <span className={styles.fileSize}>{formatSize(file.size)}</span>
              <button
                type="button"
                className={styles.remove}
                aria-label={`Убрать ${file.name}`}
                onClick={() => remove(index)}
              >
                <X size={14} />
              </button>
            </li>
          ))}
        </ul>
      ) : null}

      {shownError ? <span className={styles.error}>{shownError}</span> : null}
      {hint ? <span className={styles.hint}>{hint}</span> : null}
    </div>
  )
}