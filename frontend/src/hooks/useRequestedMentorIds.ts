import { useQuery } from '@tanstack/react-query'
import { getMyRequests } from '../api/requests'

export function useRequestedMentorIds() {
  const requests = useQuery({
    queryKey: ['requests', 'active'],
    queryFn: () => getMyRequests('active'),
  })
  return new Set(requests.data?.items.map((request) => request.mentor_id))
}
