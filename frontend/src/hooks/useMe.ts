import { useQuery } from '@tanstack/react-query'
import { getMyAccount } from '../api/auth'

export function useMe() {
  return useQuery({ queryKey: ['me'], queryFn: getMyAccount, staleTime: Infinity })
}
