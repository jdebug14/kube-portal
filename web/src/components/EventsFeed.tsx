import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '../api/client'

interface EventInvolvedObject {
  kind: string
  name: string
  namespace: string
}

interface Event {
    type: string
    reason: string
    message: string
    count: number
    first_time: string
    last_time: string
    involved_object: EventInvolvedObject
}

function EventsFeed({ namespace, involvedObjectName }: { namespace: string, involvedObjectName?: string }) {
    const url = `/api/v1/namespaces/${namespace}/events`
    + (involvedObjectName ? `?involvedObjectName=${involvedObjectName}` : '')
  const { data, isLoading, isError, error } = useQuery({
        queryKey: ['events', namespace, involvedObjectName],
        queryFn: () => apiFetch<Event[]>(url, r => r.json()),
    })
    return (
        <div>
            {isLoading && <div>Loading...</div>}
            {isError && <div>Error: {error.message}</div>}
            <h2>Events</h2>
            <ul>
                {data?.map(event => (
                    <li key={`${event.involved_object.name}-${event.reason}-${event.first_time}`}>
                        type={event.type}, name={event.involved_object.name}, reason={event.reason}, message={event.message}, count={event.count}, lastseen={event.last_time}
                    </li>
                ))}
            </ul>
        </div>
    )
}

export default EventsFeed