import { useState } from 'react';
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
    const [refetchIntervalSeconds, setRefetchIntervalSeconds] = useState(0)
    const { data, isLoading, isFetching, isError, error } = useQuery({
        queryKey: ['events', namespace, involvedObjectName],
        queryFn: () => apiFetch<Event[]>(url, r => r.json()),
        refetchInterval: refetchIntervalSeconds * 1000
    })
    return (
        <div>
            {isLoading && <div>Loading...</div>}
            {isFetching && !isLoading && <div>Refreshing...</div>}
            {isError && <div>Error: {error.message}</div>}
            <h2>Events</h2>
            Refetch interval: <select value={refetchIntervalSeconds} onChange={e => setRefetchIntervalSeconds(Number(e.target.value))}>
                <option value={0}>Never</option>
                <option value={10}>10 sec</option>
                <option value={60}>1 min</option>
                <option value={300}>5 min</option>
            </select>
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