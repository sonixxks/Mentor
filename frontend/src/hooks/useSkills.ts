import { useQuery } from '@tanstack/react-query'
import { getSkills } from '../api/profiles'

export function useSkills() {
  return useQuery({
    queryKey: ['skills'],
    queryFn: getSkills,
    staleTime: Infinity,
  })
}