import { useQuery } from '@tanstack/react-query'
import { FileText, Link2 } from 'lucide-react'
import { getWorkspaceFiles } from '../../../api/workspaces'
import type { WorkspaceDetail, WorkspaceFile } from '../../../api/workspaces'
import styles from './WorkspacePage.module.css'

function fileMeta(file: WorkspaceFile, workspace: WorkspaceDetail) {
  const author =
    file.uploaded_by === workspace.partner.id ? workspace.partner.first_name : 'добавлено тобой'
  const kind =
    file.type === 'link'
      ? 'ссылка'
      : file.file_size_kb
        ? `${file.file_size_kb.toLocaleString('ru-RU')} КБ`
        : 'файл'
  return `${kind} · ${author}`
}

export function KnowledgeBase({ workspace }: { workspace: WorkspaceDetail }) {
  const files = useQuery({
    queryKey: ['workspace', workspace.id, 'files'],
    queryFn: () => getWorkspaceFiles(workspace.id),
  })

  return (
    <section className={styles.card}>
      <h3 className={styles.cardTitle}>База знаний</h3>

      {files.isPending ? <p className={styles.muted}>Загружаем…</p> : null}

      {files.isError ? <p className={styles.error}>Не удалось загрузить материалы.</p> : null}

      {files.data?.length === 0 ? (
        <p className={styles.muted}>
          Материалов пока нет. Наставник добавит сюда файлы и полезные ссылки.
        </p>
      ) : null}

      {files.data?.map((file) => {
        const Icon = file.type === 'link' ? Link2 : FileText
        return (
          <a
            key={file.id}
            href={file.file_url}
            target="_blank"
            rel="noreferrer"
            className={styles.file}
          >
            <span className={styles.fileIcon}>
              <Icon size={18} />
            </span>
            <span className={styles.fileInfo}>
              <span className={styles.fileName}>{file.file_name}</span>
              <span className={styles.muted}>{fileMeta(file, workspace)}</span>
            </span>
          </a>
        )
      })}
    </section>
  )
}
