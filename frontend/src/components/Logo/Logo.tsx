import { Link } from 'react-router-dom'
import logo from '../../assets/logo.svg'
import styles from './Logo.module.css'

type LogoProps = {
  to?: string
}

export function Logo({ to = '/' }: LogoProps) {
  return (
    <Link to={to} className={styles.logo}>
      <img src={logo} alt="" className={styles.mark} />
      <span className={styles.text}>Cosmomentor</span>
    </Link>
  )
}