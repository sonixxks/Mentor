import styles from './CuratorHelpCard.module.css'

export function CuratorHelpCard() {
  return (
    <div className={styles.card}>
      <span className={styles.title}>Не можешь выбрать?</span>
      <p className={styles.text}>Куратор поможет подобрать наставника вручную.</p>
      <button type="button" className={styles.action}>
        Запросить помощь
      </button>
    </div>
  )
}
