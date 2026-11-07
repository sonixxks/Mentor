export type Role = 'student' | 'mentor' | 'curator' | 'admin'

/** Кураторов создаёт администратор. */
export type PublicRole = 'student' | 'mentor'